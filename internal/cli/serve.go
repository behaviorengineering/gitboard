package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/llm"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/internal/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/internal/triage"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

const (
	defaultAddr = "127.0.0.1:1325"
	// shutdownBudget stays under process-compose shutdown.timeout_seconds (10)
	// so the port is freed before a watch restart binds again.
	shutdownBudget = 8 * time.Second
	// observabilityFlushBudget bounds the trace exporter flush on exit.
	observabilityFlushBudget = 5 * time.Second
)

type serveFlags struct {
	path              string
	projects          string
	addr              string
	allowNonLocalhost bool
}

func newServeCmd() *cobra.Command {
	var f serveFlags
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the local dashboard (default 127.0.0.1:1325); creates config if missing",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd.Context(), f)
		},
	}
	cmd.Flags().StringVar(&f.path, "config", config.DefaultPath(), "config file path, created if missing (GITBOARD_CONFIG, then ~/.config/gitboard/config.yaml)")
	cmd.Flags().StringVar(&f.projects, "projects", "", "legacy alias for --config")
	cmd.Flags().StringVar(&f.addr, "addr", defaultAddr, "listen address (localhost only)")
	cmd.Flags().BoolVar(&f.allowNonLocalhost, "allow-non-localhost", false, "allow bind on non-loopback")
	return cmd
}

func runServe(ctx context.Context, f serveFlags) error {
	path := f.path
	if strings.TrimSpace(f.projects) != "" {
		path = f.projects
	}
	if !f.allowNonLocalhost && !strings.HasPrefix(f.addr, "127.0.0.1:") && !strings.HasPrefix(f.addr, "localhost:") {
		return errors.New("refuse to bind outside localhost (use --allow-non-localhost to override)")
	}
	created, err := config.Init(path)
	if err != nil {
		return fmt.Errorf("ensure config: %w", err)
	}
	if created {
		log.Printf("gitboard: created config %s (use Manage or gitboard sync to add projects)", path)
	}
	doc, err := config.Load(path)
	if err != nil {
		return err
	}
	tp, err := initObservability(doc)
	if err != nil {
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), observabilityFlushBudget)
		defer cancel()
		if err := observability.Shutdown(flushCtx, tp); err != nil {
			log.Printf("gitboard: observability shutdown: %v", err)
		}
	}()

	handler, err := newDashboardHandler(f.addr, path, doc)
	if err != nil {
		return err
	}
	log.Printf("gitboard: %s (%d projects, config %s, poll %ds, heads cache %ds, merged cache %ds, fetch cache %ds)",
		f.addr, len(doc.Projects), path, doc.EffectivePollSeconds(),
		doc.EffectiveHeadsSeconds(), doc.EffectiveMergedSeconds(), doc.EffectiveFetchSeconds())
	log.Printf("gitboard: uses gh and glab; local roots=%d; AI triage via llm in config; agents %s",
		len(doc.Local.Roots), config.AgentsDir())
	return listenUntilDone(ctx, f.addr, handler)
}

func initObservability(doc config.File) (*trace.TracerProvider, error) {
	oi := doc.EffectiveOpenInference()
	oiEnabled := oi.Enabled != nil && *oi.Enabled
	var dumpDir, otlpEndpoint string
	if oiEnabled {
		otlpEndpoint = oi.Endpoint
		if oi.FailureDump.Enabled != nil && *oi.FailureDump.Enabled {
			dumpDir = oi.FailureDump.Dir
		}
	}
	tp, err := observability.Init(observability.InitConfig{
		ServiceName:            oi.ServiceName,
		OTLPEndpoint:           otlpEndpoint,
		FailureDumpDir:         dumpDir,
		FailureDumpMaxAgeHours: oi.FailureDump.MaxAgeHours,
		FailureDumpMaxFiles:    oi.FailureDump.MaxFiles,
	})
	if err != nil {
		return nil, fmt.Errorf("observability: %w", err)
	}
	return tp, nil
}

func newDashboardHandler(addr, path string, doc config.File) (http.Handler, error) {
	run := cliexec.New()
	run.Timeout = 120 * time.Second
	dash := dashboard.New(
		remotegit.NewGitHub(run),
		remotegit.NewGitLab(run),
		remotegit.NewAzureDevOps(run),
		remotegit.NewBitbucket(run),
		localgit.NewInspector(run),
	)
	llmClient := llm.New(doc.EffectiveLLM())
	prune, err := pruneagent.New(config.AgentsDir(), run, llmClient)
	if err != nil {
		return nil, fmt.Errorf("prune agent: %w", err)
	}
	return server.NewMux(server.Options{
		Addr:        addr,
		ConfigPath:  path,
		Doc:         doc,
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		Triage:      &triage.Analyzer{LLM: llmClient},
		Prune:       prune,
		PollSeconds: doc.EffectivePollSeconds(),
		Log:         observability.NewLogger().With("cmd", "serve"),
	}), nil
}

// listenUntilDone serves until the listener fails or ctx is cancelled (SIGINT or
// SIGTERM at the entrypoint), then shuts down within shutdownBudget.
func listenUntilDone(ctx context.Context, addr string, handler http.Handler) error {
	// In-flight handlers derive from r.Context(), so cancelling baseCtx stops them.
	baseCtx, baseCancel := context.WithCancel(ctx)
	defer baseCancel()
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       120 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return baseCtx
		},
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		log.Printf("gitboard: shutting down (%v)", context.Cause(ctx))
		baseCancel()
		return shutdown(ctx, srv, errCh)
	}
}

func shutdown(ctx context.Context, srv *http.Server, errCh <-chan error) error {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownBudget)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("gitboard: graceful shutdown: %v", err)
		if closeErr := srv.Close(); closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
			log.Printf("gitboard: force close: %v", closeErr)
		}
	}
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-time.After(2 * time.Second):
		log.Printf("gitboard: listener exit still pending; continuing shutdown")
	}
	return nil
}
