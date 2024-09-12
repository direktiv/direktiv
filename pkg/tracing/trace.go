package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TraceGWHTTPRequest(ctx context.Context, r *http.Request, instrumentationName string) func() {
	tp := otel.GetTracerProvider()
	tr := tp.Tracer(instrumentationName)
	ctx, span := tr.Start(ctx, "starting gateway request to flow", trace.WithSpanKind(trace.SpanKindClient))

	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, &httpCarrier{
		r: r,
	})

	return func() { span.End() }
}

type httpCarrier struct {
	r *http.Request
}

func (c *httpCarrier) Get(key string) string {
	return c.r.Header.Get(key)
}

// nolint:canonicalheader
func (c *httpCarrier) Keys() []string {
	return c.r.Header.Values("oteltmckeys")
}

// nolint:canonicalheader
func (c *httpCarrier) Set(key, val string) {
	prev := c.Get(key)
	if prev == "" {
		c.r.Header.Add("oteltmckeys", key)
	}
	c.r.Header.Set(key, val)
}
