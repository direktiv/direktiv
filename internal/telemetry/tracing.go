package telemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	OtelServiceName         = "direktiv"
	TracingObjectIdentifier = "tracing"
)

func InitOpenTelemetry(ctx context.Context, otelURL string) error {
	// skip telemetry
	if otelURL == "" {
		slog.Info("telemetry not configured")

		// create dummy doing nothing
		provider := tracesdk.NewTracerProvider()
		otel.SetTracerProvider(provider)

		return nil
	}

	slog.Info("initializing opentelemetry")

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(otelURL),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithReconnectionPeriod(time.Second*10),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			MaxInterval:     3 * time.Second,
			InitialInterval: 1 * time.Second,
			MaxElapsedTime:  1 * time.Minute,
		}),
	)
	if err != nil {
		slog.Error("opentelemetry setup failed", slog.Any("error", err))
		return err
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(OtelServiceName),
	)

	provider := tracesdk.NewTracerProvider(
		tracesdk.WithResource(resource),
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.AlwaysSample())),
		tracesdk.WithBatcher(exporter, tracesdk.WithBatchTimeout(time.Second)),
	)
	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, // W3C Trace Context format; https://www.w3.org/TR/trace-context/
			propagation.Baggage{},
		),
	)

	return nil
}
