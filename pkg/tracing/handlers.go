package tracing

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func MuxMiddleware(imName string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the trace context from HTTP request headers
			prop := otel.GetTextMapPropagator()
			ctx := prop.Extract(r.Context(), &httpCarrier{r: r})

			tp := otel.GetTracerProvider()
			tr := tp.Tracer(imName)

			// Determine the route name from mux (if available)
			route := "apiv2"
			if mux.CurrentRoute(r) != nil {
				route = mux.CurrentRoute(r).GetName()
			}

			// Start a new trace span
			ctx, span := tr.Start(ctx, route, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			// Pass the context to the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ChiTelemetryMiddleware is the middleware for chi-based routers to handle tracing.
func ChiTelemetryMiddleware(imName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from HTTP headers
			prop := otel.GetTextMapPropagator()
			ctx := prop.Extract(r.Context(), &httpCarrier{r: r})

			// Create tracer and start span
			tr := otel.Tracer(imName)
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			ctx, span := tr.Start(ctx, routePattern, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			// Pass the context with the span to the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TraceHTTPRequest starts a client trace for outgoing HTTP requests.
// The span name is customizable to allow for more flexibility.
func TraceHTTPRequest(ctx context.Context, r *http.Request, spanName string) (cleanup func()) {
	tr := otel.Tracer(instrumentationName)
	ctx, span := tr.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))

	// Inject the trace context into the outgoing HTTP request headers
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, &httpCarrier{r: r})

	return func() { span.End() }
}

// TraceGWHTTPRequest starts a trace for gateway-related HTTP requests.
func TraceGWHTTPRequest(ctx context.Context, r *http.Request, instrumentationName string) (cleanup func()) {
	return TraceHTTPRequest(ctx, r, "gateway-"+instrumentationName)
}

// httpCarrier implements the TextMapCarrier interface for HTTP requests.
type httpCarrier struct {
	r *http.Request
}

// Get retrieves a header value by key.
func (c *httpCarrier) Get(key string) string {
	return c.r.Header.Get(key)
}

// Keys retrieves all header keys that have trace-related information.
func (c *httpCarrier) Keys() []string {
	keys := make([]string, 0, len(c.r.Header))
	for k := range c.r.Header {
		keys = append(keys, k)
	}

	return keys
}

// Set sets a header value by key.
func (c *httpCarrier) Set(key, val string) {
	c.r.Header.Set(key, val)
}
