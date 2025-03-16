package tracing

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/version"
	"github.com/go-chi/chi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otlpmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

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

			defer func() {
				if requestCounter == nil || requestDuration == nil {
					return
				}
				duration := time.Since(startTime).Seconds()

				// Record Metrics
				requestCounter.Add(ctx, 1, otlpmetric.WithAttributes(
					attribute.String("http.method", method),
					attribute.String("http.route", routePattern),
				))

				requestDuration.Record(ctx, duration, otlpmetric.WithAttributes(
					attribute.String("http.method", method),
					attribute.String("http.route", routePattern),
				))

				span.End()
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
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
