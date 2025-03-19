package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "direktiv"
)

var Tracer oteltrace.Tracer

func InitOpenTelemetry(ctx context.Context, otelURL string) (*trace.TracerProvider, error) {
	// skip telemetry
	if otelURL == "" {
		slog.Info("telemetry not configured")

		// create dummy doing nothing
		provider := trace.NewTracerProvider()
		otel.SetTracerProvider(provider)
		Tracer = otel.Tracer(serviceName)

		return trace.NewTracerProvider(), nil
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
		return nil, err
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	provider := trace.NewTracerProvider(
		trace.WithResource(resource),
		trace.WithSampler(trace.ParentBased(trace.AlwaysSample())),
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, // W3C Trace Context format; https://www.w3.org/TR/trace-context/
			propagation.Baggage{},
		),
	)

	// create tracer
	Tracer = otel.Tracer(serviceName)

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

func ReportError(span oteltrace.Span, err error) {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}

// GetContextFromRequest looks for the tracing parent in http header
func GetContextFromRequest(r *http.Request) context.Context {
	propagator := propagation.TraceContext{}
	ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	return ctx
}

// FromTraceParent create a context based on a traceparent string
func FromTraceParent(ctx context.Context, traceparent string) context.Context {
	mc := make(propagation.MapCarrier)
	mc.Set("traceparent", traceparent)

	tc := propagation.TraceContext{}
	return tc.Extract(ctx, mc)
}

// TraceParent returns the traceparent value as string from context
func TraceParent(ctx context.Context) string {
	mc := make(propagation.MapCarrier)
	tc := propagation.TraceContext{}
	tc.Inject(ctx, mc)

	return mc.Get("traceparent")
}
