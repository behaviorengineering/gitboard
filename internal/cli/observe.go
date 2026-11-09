package cli

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/behaviorengineering/olly/pkg/dump"
	ollyhttp "github.com/behaviorengineering/olly/pkg/http"
	"github.com/behaviorengineering/olly/pkg/olly"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/observability"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
)

// observabilityFlushBudget bounds the trace flush on exit.
const observabilityFlushBudget = 5 * time.Second

// tracing is the olly setup for one process: the config handed to olly.Init
// and the service name for spans.
type tracing struct {
	cfg olly.Config
}

// newTracing maps the effective openinference settings onto olly. The OTLP
// endpoint comes from GITBOARD_OPENINFERENCE_ENDPOINT, POLYPUS_OTLP_ENDPOINT
// or openinference.endpoint; when all are empty olly falls back to
// OTEL_EXPORTER_OTLP_TRACES_ENDPOINT and OTEL_EXPORTER_OTLP_ENDPOINT. An
// http:// or https:// endpoint is sent OTLP over HTTP; a bare host:port uses
// gRPC unless OTEL_EXPORTER_OTLP_PROTOCOL says otherwise. withDumps turns on
// the failure-dump processor when the config enables it.
func newTracing(doc boardconfig.File, log observability.Logger, withDumps bool) tracing {
	oi := config.EffectiveOpenInference(doc)
	cfg := olly.Config{
		Enabled:      oi.Enabled != nil && *oi.Enabled,
		ServiceName:  oi.ServiceName,
		OTLPEndpoint: oi.Endpoint,
		Diagnostics:  logDiagnostics{log: log},
	}
	if strings.Contains(oi.Endpoint, "://") {
		cfg.Protocol = olly.ProtocolHTTP
	}
	if withDumps && oi.FailureDump.Enabled != nil && *oi.FailureDump.Enabled {
		cfg.Dump = dump.Config{
			Dir:         oi.FailureDump.Dir,
			MaxAgeHours: oi.FailureDump.MaxAgeHours,
			MaxFiles:    oi.FailureDump.MaxFiles,
		}
	}
	return tracing{cfg: cfg}
}

// start installs the global tracer provider and returns stop, which flushes
// and shuts it down within observabilityFlushBudget.
func (t tracing) start(ctx context.Context, log observability.Logger) (stop func(), err error) {
	shutdown, err := olly.Init(t.cfg)
	if err != nil {
		return nil, fmt.Errorf("observability: %w", err)
	}
	if t.cfg.Dump.Dir != "" {
		log.With("dir", t.cfg.Dump.Dir).Infof("failure dumps on (error-only traces)")
	}
	return func() {
		flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), observabilityFlushBudget)
		defer cancel()
		if err := shutdown(flushCtx); err != nil {
			log.Errorf("observability shutdown: %v", err)
		}
	}, nil
}

// httpClient returns an outbound client whose requests carry trace context
// and record client spans.
func (t tracing) httpClient() *http.Client {
	return ollyhttp.Client(&http.Client{}, t.serviceName())
}

func (t tracing) serviceName() string {
	if t.cfg.ServiceName == "" {
		return "gitboard"
	}
	return t.cfg.ServiceName
}

// logDiagnostics sends olly's dump messages to the process logger.
type logDiagnostics struct {
	log observability.Logger
}

func (d logDiagnostics) Printf(format string, args ...any) {
	d.log.Infof(format, args...)
}
