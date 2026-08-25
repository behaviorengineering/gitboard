// Package observability bootstraps OTEL and writes error-only inference dumps
// (same pattern as content-pipelines).
package observability

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitConfig bootstraps the process tracer provider.
type InitConfig struct {
	ServiceName            string
	OTLPEndpoint           string
	FailureDumpDir         string
	FailureDumpMaxAgeHours int
	FailureDumpMaxFiles    int
}

// Init installs the global TracerProvider. OTLP is optional; failure dumps
// write full traces when a root span ends with ERROR.
func Init(cfg InitConfig) (*sdktrace.TracerProvider, error) {
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "gitboard"
	}
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1.0)),
	}
	if cfg.FailureDumpDir != "" {
		opts = append(opts, sdktrace.WithSpanProcessor(newFailureDumpProcessor(
			cfg.FailureDumpDir,
			cfg.FailureDumpMaxAgeHours,
			cfg.FailureDumpMaxFiles,
		)))
	}
	if cfg.OTLPEndpoint != "" {
		exporter, exportErr := otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if exportErr != nil {
			return nil, fmt.Errorf("otlp exporter: %w", exportErr)
		}
		opts = append(opts, sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(2*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		))
	}
	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp, nil
}

// Shutdown flushes and shuts down the provider.
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

func logf(format string, args ...any) {
	log.Printf("gitboard/observability: "+format, args...)
}
