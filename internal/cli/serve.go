package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/behaviorengineering/gitboard/internal/ai/llm"
	"github.com/behaviorengineering/gitboard/internal/ai/pruneagent"
	"github.com/behaviorengineering/gitboard/internal/ai/triage"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/clock"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

const (
	defaultAddr = "127.0.0.1:1325"
	// shutdownBudget stays under process-compose shutdown.timeout_seconds (10)
	// so the port is freed before a watch restart binds again.
	shutdownBudget = 8 * time.Second
	// listenerExitWait bounds the wait for ListenAndServe to return after
	// Shutdown.
	listenerExitWait = 2 * time.Second
	// HTTP server limits: header read, whole request read, and idle
	// keep-alive.
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 60 * time.Second
	idleTimeout       = 120 * time.Second
	// launchTokenBytes is the size of the random API token generated with
	// --allow-non-localhost.
	launchTokenBytes = 32
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
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd.Context(), f)
		},
	}
	cmd.Flags().StringVar(&f.path, "config", config.DefaultPath(), configFlagHelp+"; created if missing")
	cmd.Flags().StringVar(&f.projects, "projects", "", "legacy alias for --config")
	cmd.Flags().StringVar(&f.addr, "addr", defaultAddr, "listen address (localhost only)")
	cmd.Flags().BoolVar(&f.allowNonLocalhost, "allow-non-localhost", false,
		"allow bind on non-loopback; mutating API calls then need the token printed at startup")
	return cmd
}

func runServe(ctx context.Context, f serveFlags) error {
	log := observability.NewLogger().With("cmd", "serve")
	path := f.path
	if strings.TrimSpace(f.projects) != "" {
		path = f.projects
	}
	if !f.allowNonLocalhost && !strings.HasPrefix(f.addr, "127.0.0.1:") && !strings.HasPrefix(f.addr, "localhost:") {
		return usage(errors.New("refuse to bind outside localhost (use --allow-non-localhost to override)"))
	}
	created, err := config.Init(path)
	if err != nil {
		return fmt.Errorf("ensure config: %w", err)
	}
	if created {
		log.With("path", path).Infof("created config (use Manage or gitboard sync to add projects)")
	}
	doc, err := loadConfig(path)
	if err != nil {
		return err
	}
	tr := newTracing(doc, log, true)
	stop, err := tr.start(ctx, log)
	if err != nil {
		return err
	}
	defer stop()

	token := ""
	if f.allowNonLocalhost {
		if token, err = newLaunchToken(); err != nil {
			return err
		}
		log.Infof("API token for this launch (send it in %s; the dashboard reads it from /api/meta): %s", server.TokenHeader, token)
	}
	handler, err := newDashboardHandler(dashboardWiring{addr: f.addr, path: path, token: token, doc: doc, tracing: tr, log: log})
	if err != nil {
		return err
	}
	log.With("addr", f.addr, "projects", len(doc.Projects), "config", path).
		Infof("listening (poll %ds, heads cache %ds, merged cache %ds, fetch cache %ds)",
			config.EffectivePollSeconds(doc), doc.EffectiveHeadsSeconds(), doc.EffectiveMergedSeconds(), doc.EffectiveFetchSeconds())
	log.Infof("forges: gh (GitHub), glab (GitLab), az (Azure DevOps) and the Bitbucket REST API; local roots=%d; AI triage via llm in config; agents %s",
		len(doc.Local.Roots), config.AgentsDir())
	return listenUntilDone(ctx, newHTTPServer(f.addr, handler), log)
}

// newLaunchToken returns a random hex token for one serve launch.
func newLaunchToken() (string, error) {
	b := make([]byte, launchTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate API token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// dashboardWiring is what newDashboardHandler needs from serve.
type dashboardWiring struct {
	addr    string
	path    string
	token   string
	doc     boardconfig.File
	tracing tracing
	log     observability.Logger
}

func newDashboardHandler(w dashboardWiring) (http.Handler, error) {
	run := newRunner()
	now := clock.Clock(time.Now)
	cfg := dashboardConfig(run, now)
	cfg.Local = dashboard.NewLocalGit(run)
	cfg.Mutations = localgit.NewMutationCoordinator(localgit.DefaultOSLocker())
	dash := cfg.CreateService()
	llmClient := llm.New(config.EffectiveLLM(w.doc), w.tracing.httpClient())
	prune, err := pruneagent.Config{
		AgentsDir: config.AgentsDir(),
		Exec:      run,
		LLM:       llmClient,
		Mutations: cfg.Mutations,
		Clock:     now,
	}.CreateService()
	if err != nil {
		return nil, fmt.Errorf("prune agent: %w", err)
	}
	return server.NewMux(server.Options{
		Addr:        w.addr,
		Token:       w.token,
		ServiceName: w.tracing.serviceName(),
		ConfigPath:  w.path,
		Doc:         w.doc,
		Dash:        dash,
		Commands:    dashboard.NewCommands(dash),
		Triage:      &triage.Analyzer{LLM: llmClient},
		Prune:       prune,
		PollSeconds: config.EffectivePollSeconds(w.doc),
		Log:         w.log,
	}), nil
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// listenUntilDone serves until the listener fails or ctx is cancelled (SIGINT or
// SIGTERM at the entrypoint), then shuts down within shutdownBudget.
func listenUntilDone(ctx context.Context, srv *http.Server, log observability.Logger) error {
	// In-flight handlers derive from r.Context(), so cancelling baseCtx stops them.
	baseCtx, baseCancel := context.WithCancel(ctx)
	defer baseCancel()
	srv.BaseContext = func(net.Listener) context.Context {
		return baseCtx
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
		log.Infof("shutting down (%v)", context.Cause(ctx))
		baseCancel()
		return shutdown(ctx, srv, errCh, log)
	}
}

func shutdown(ctx context.Context, srv *http.Server, errCh <-chan error, log observability.Logger) error {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownBudget)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("graceful shutdown: %v", err)
		if closeErr := srv.Close(); closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
			log.Errorf("force close: %v", closeErr)
		}
	}
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-time.After(listenerExitWait):
		log.Infof("listener exit still pending; continuing shutdown")
	}
	return nil
}
