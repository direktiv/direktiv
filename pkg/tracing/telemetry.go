package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/middlewares"
	"github.com/direktiv/direktiv/pkg/version"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otlpmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var (
	instrumentationName string
	meter               otlpmetric.Meter
	requestCounter      otlpmetric.Int64Counter
	requestDuration     otlpmetric.Float64Histogram
)

func SetInstrumentationName(name string) {
	instrumentationName = name
}

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
		otlpgrpc.WithDialOption(grpc.WithBlock()), // nolint:staticcheck
	)
	// Set up OTLP Metric Exporter
	slog.Debug("Creating OTLP metric exporter.", "endpoint", addr)
	metricExporter, err := otlpmetricgrpc.New(cirCtx, otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithEndpoint(addr))
	if err != nil {
		slog.Error("Failed to create OTLP metric exporter.", "error", err)
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Create new OpenTelemetry metric provider
	slog.Debug("Setting up OpenTelemetry metric provider.")
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)

	otel.SetMeterProvider(meterProvider)
	meter = meterProvider.Meter(instrumentationName)

	// Define metrics
	requestCounter, err = meter.Int64Counter(
		"http.server.requests",
		otlpmetric.WithDescription("Total number of HTTP requests received"),
	)
	if err != nil {
		slog.Error("Failed to create requestCounter metric.", "error", err)
		return nil, fmt.Errorf("failed to create requestCounter metric: %w", err)
	}

	requestDuration, err = meter.Float64Histogram(
		"http.server.duration",
		otlpmetric.WithDescription("Duration of HTTP requests"),
	)
	if err != nil {
		slog.Error("Failed to create requestDuration metric.", "error", err)
		return nil, fmt.Errorf("failed to create requestDuration metric: %w", err)
	}

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

	sampler := sdktrace.AlwaysSample()

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
		return entrypointOtelMiddleware(imName, h)
	})

	slog.Debug("Telemetry initialization completed.")

	return telemetryWaiter(cirCtx, tp, bsp), nil
}

// telemetryWaiter ensures all telemetry data is flushed and the provider is shut down gracefully.
func telemetryWaiter(cirCtx context.Context, tp *sdktrace.TracerProvider, bsp sdktrace.SpanProcessor) func() {
	return func() {
		ctx, cancel := context.WithTimeout(cirCtx, 25*time.Second)
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

// entrypointOtelMiddleware injects trace context into the request and starts a new span.
func entrypointOtelMiddleware(imName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		ctx := r.Context()
		parentSpan := trace.SpanFromContext(ctx)
		var span trace.Span
		tracer := otel.Tracer(instrumentationName)

		route := extractRoute(r)
		method := r.Method
		namespace := extractNamespace(r)
		apiVersion := version.Version

		if parentSpan.SpanContext().IsValid() {
			ctx, span = tracer.Start(ctx, fmt.Sprintf("%s-child:%s", imName, route), trace.WithSpanKind(trace.SpanKindInternal))
		} else {
			ctx, span = tracer.Start(ctx, fmt.Sprintf("%s-root:%s", imName, route))
		}

		span.SetAttributes(
			attribute.String("http.route", route),
			attribute.String("http.method", method),
			attribute.String("namespace", namespace),
			attribute.String("api.version", apiVersion),
			attribute.String("instance.manager", imName),
		)

		// Wrap ResponseWriter to capture response status
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			duration := time.Since(startTime).Seconds()

			// Record Metrics
			requestCounter.Add(ctx, 1, otlpmetric.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.route", route),
				attribute.Int("http.status", rw.statusCode),
			))

			requestDuration.Record(ctx, duration, otlpmetric.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.route", route),
				attribute.Int("http.status", rw.statusCode),
			))

			span.End()
		}()

		r = r.WithContext(ctx)
		next.ServeHTTP(rw, r)
	})
}

// responseWriter is a wrapper around http.ResponseWriter to capture status codes.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// extractRoute attempts to extract the API route from the request.
func extractRoute(r *http.Request) string {
	if chiCtx := chi.RouteContext(r.Context()); chiCtx != nil && chiCtx.RoutePath != "" {
		return chiCtx.RoutePath
	}

	return r.URL.Path
}

// extractNamespace attempts to extract the namespace from chi params or the URL path.
func extractNamespace(r *http.Request) string {
	if chiCtx := chi.RouteContext(r.Context()); chiCtx != nil {
		namespace := chi.URLParam(r, "namespace")
		if namespace != "" {
			return namespace
		}
	}
	pathSegments := splitURLPath(r.URL.Path)
	for i, segment := range pathSegments {
		if segment == "namespaces" && i+1 < len(pathSegments) {
			return pathSegments[i+1]
		}
	}

	return "unknown"
}

func splitURLPath(path string) []string {
	cleanPath := strings.Trim(path, "/")
	return strings.Split(cleanPath, "/")
}
