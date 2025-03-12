package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	serviceName = "direktiv"
)

func InitOpenTelemetry(ctx context.Context, otelURL string) (*trace.TracerProvider, error) {
	slog.Info("initializing opentelemetry")
	fmt.Println(otelURL)
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(otelURL),
		otlptracegrpc.WithInsecure(),
	)

	// WithReconnectionPeriod
	// WithRetry

	if err != nil {
		slog.Error("opentelemetry setup failed", slog.Any("error", err))
		return nil, err
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	provider := trace.NewTracerProvider(
		trace.WithResource(resource),
		trace.WithSampler(trace.ParentBased(trace.AlwaysSample())),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter)),
		trace.WithBatcher(exporter),
	)
	// otel.SetTracerProvider(provider)

	// tp, err := newTraceProvider(ctx)
	// fmt.Println(err)
	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, // W3C Trace Context format; https://www.w3.org/TR/trace-context/
			propagation.Baggage{},
		),
	)

	_, span := provider.Tracer("jens").Start(ctx, "ssss")
	span.End()

	return provider, nil
}

func newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}
