// nolint:unused
package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
)

var (
	instrumentationName string
	meter               otlpmetric.Meter
	requestCounter      otlpmetric.Int64Counter
	requestDuration     otlpmetric.Float64Histogram
)

const flushingTimeoutSeconds = 25

// InitTelemetry initializes tracing with OTLP, resource, and tracer provider setup.
func InitTelemetry(cirCtx context.Context, addr string, svcName, imName string) (func(), error) {
	slog.Debug("initializing telemetry.", slog.String("instrumentationName", imName))
	instrumentationName = imName

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if addr == "" {
		slog.Warn("no OTLP address provided. Telemetry will not be exported.")
		return func() {}, nil
	}

	slog.Debug("creating OTLP gRPC client.", slog.String("endpoint", addr))
	driver := otlpgrpc.NewClient(
		otlpgrpc.WithEndpoint(addr),
		otlpgrpc.WithInsecure(),
	)

	// Retry strategy for OTLP Metric Exporter
	maxRetries := 3
	var metricExporter *otlpmetricgrpc.Exporter
	var err error
	for attempt := range maxRetries {
		metricExporter, err = otlpmetricgrpc.New(cirCtx, otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithEndpoint(addr))
		if err == nil {
			break
		}
		slog.Warn("failed to create OTLP metric exporter, retrying...", slog.Int("attempt", attempt+1), slog.Any("error", err))
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
	if err != nil {
		slog.Error("failed to create OTLP metric exporter after retries.", slog.Any("error", err))
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	slog.Debug("setting up OpenTelemetry metric provider.")
	res, err := resource.New(cirCtx, resource.WithAttributes(semconv.ServiceNameKey.String(svcName)))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	meter = meterProvider.Meter(instrumentationName)

	requestCounter, err = meter.Int64Counter("http.server.requests", otlpmetric.WithDescription("Total number of HTTP requests received"))
	if err != nil {
		return nil, fmt.Errorf("failed to create requestCounter metric: %w", err)
	}

	requestDuration, err = meter.Float64Histogram("http.server.duration", otlpmetric.WithDescription("Duration of HTTP requests"))
	if err != nil {
		return nil, fmt.Errorf("failed to create requestDuration metric: %w", err)
	}

	// Retry strategy for OTLP Exporter
	var exp *otlp.Exporter
	for attempt := range maxRetries {
		exp, err = otlp.New(cirCtx, driver)
		if err == nil {
			break
		}
		slog.Warn("failed to create OTLP exporter, retrying...", slog.Int("attempt", attempt+1), slog.Any("error", err))
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
	if err != nil {
		slog.Error("failed to create OTLP exporter after retries.", slog.Any("error", err))
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.AlwaysSample())
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	slog.Debug("setting tracer provider.")
	otel.SetTracerProvider(tp)

	return telemetryWaiter(cirCtx, tp, bsp), nil
}

// telemetryWaiter ensures all telemetry data is flushed and the provider is shut down gracefully.
func telemetryWaiter(cirCtx context.Context, tp *sdktrace.TracerProvider, bsp sdktrace.SpanProcessor) func() {
	return func() {
		ctx, cancel := context.WithTimeout(cirCtx, flushingTimeoutSeconds*time.Second)
		defer cancel()

		// Force flush to export all remaining telemetry data
		slog.Info("flushing telemetry data before shutdown.")
		if err := bsp.ForceFlush(ctx); err != nil {
			slog.Error("failed to flush telemetry data.", "error", err)
		}

		// Shut down the tracer provider
		slog.Info("shutting down telemetry.")
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown telemetry.", "error", err)
		}
	}
}

// OtelMiddleware injects trace context into the request and starts a new span.
func OtelMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			ctx := r.Context()
			parentSpan := trace.SpanFromContext(ctx)
			var span trace.Span
			tracer := otel.Tracer(instrumentationName)

			routePattern := getRoutePattern(r)
			method := r.Method
			namespace := extractNamespace(r)
			apiVersion := version.Version

			if parentSpan.SpanContext().IsValid() {
				ctx, span = tracer.Start(ctx, fmt.Sprintf("%s-child:%s", instrumentationName, routePattern), trace.WithSpanKind(trace.SpanKindInternal))
			} else {
				ctx, span = tracer.Start(ctx, fmt.Sprintf("%s-root:%s", instrumentationName, routePattern))
			}

			span.SetAttributes(
				attribute.String("http.route", routePattern),
				attribute.String("http.method", method),
				attribute.String("namespace", namespace),
				attribute.String("api.version", apiVersion),
				attribute.String("instance.manager", instrumentationName),
			)

			// Wrap ResponseWriter to capture response status
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			defer func() {
				if requestCounter == nil || requestDuration == nil {
					return
				}
				duration := time.Since(startTime).Seconds()

				// Record Metrics
				requestCounter.Add(ctx, 1, otlpmetric.WithAttributes(
					attribute.String("http.method", method),
					attribute.String("http.route", routePattern),
					attribute.Int("http.status", rw.statusCode),
				))

				requestDuration.Record(ctx, duration, otlpmetric.WithAttributes(
					attribute.String("http.method", method),
					attribute.String("http.route", routePattern),
					attribute.Int("http.status", rw.statusCode),
				))

				span.End()
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(rw, r)
		})
	}
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

// getRoutePattern extracts the route pattern from Chi's RouteContext.
func getRoutePattern(r *http.Request) string {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil || len(routeContext.RoutePatterns) == 0 {
		return "unknown"
	}
	// Return the last matched route pattern (most specific)
	return routeContext.RoutePatterns[len(routeContext.RoutePatterns)-1]
}

func splitURLPath(path string) []string {
	cleanPath := strings.Trim(path, "/")
	return strings.Split(cleanPath, "/")
}
