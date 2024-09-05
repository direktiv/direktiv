package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/middlewares"
	"go.opentelemetry.io/otel"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"google.golang.org/grpc"
)

var instrumentationName string

// InitTelemetry initializes tracing with OTLP, resource, and tracer provider setup.
func InitTelemetry(cirCtx context.Context, addr string, svcName, imName string) (func(), error) {
	slog.Debug("Initializing telemetry.", "instrumentationName", imName)
	instrumentationName = imName

	// Setup context propagation format
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if addr == "" {
		slog.Warn("No OTLP address provided. Telemetry will not be exported.")
		return func() {}, nil
	}

	// Setup OTLP exporter
	slog.Debug("Creating OTLP gRPC client.", "endpoint", addr)
	driver := otlpgrpc.NewClient(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(addr),
		otlpgrpc.WithDialOption(grpc.WithBlock()),
	)

	slog.Debug("Setting up OTLP exporter.")
	exp, err := otlp.New(cirCtx, driver)
	if err != nil {
		slog.Error("Failed to create OTLP exporter.", "error", err)
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Setup resource with service name
	slog.Debug("Creating resource with service name.", "serviceName", svcName)
	res, err := resource.New(cirCtx, resource.WithAttributes(semconv.ServiceNameKey.String(svcName)))
	if err != nil {
		slog.Error("Failed to create resource.", "error", err)
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Choose a sampler based on an environment variable or default to AlwaysSample
	sampler := sdktrace.AlwaysSample() // You could configure this based on env variables

	// Set up batch span processor and tracer provider
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	slog.Debug("Setting tracer provider.")
	otel.SetTracerProvider(tp)

	// Register HTTP telemetry middleware
	slog.Debug("Registering HTTP telemetry middleware.")
	middlewares.RegisterHTTPMiddleware(func(h http.Handler) http.Handler {
		return otelMiddleware(imName, h)
	})

	slog.Debug("Telemetry initialization completed.")

	return telemetryWaiter(tp, bsp), nil
}

// telemetryWaiter ensures all telemetry data is flushed and the provider is shut down gracefully.
func telemetryWaiter(tp *sdktrace.TracerProvider, bsp sdktrace.SpanProcessor) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		// Force flush to export all remaining telemetry data
		slog.Info("Flushing telemetry data before shutdown.")
		if err := bsp.ForceFlush(ctx); err != nil {
			slog.Error("Failed to flush telemetry data.", "error", err)
		}

		// Shut down the tracer provider
		slog.Info("Shutting down telemetry.")
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown telemetry.", "error", err)
		}
	}
}

// otelMiddleware injects trace context into the request and starts a new span.
func otelMiddleware(imName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tr := otel.Tracer(instrumentationName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s-request", imName))
		defer span.End()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
