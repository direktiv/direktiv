package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
)

type GenericTelemetryCarrier struct {
	Trace map[string]string
}

func (c *GenericTelemetryCarrier) Get(key string) string {
	v := c.Trace[key]
	return v
}

func (c *GenericTelemetryCarrier) Keys() []string {
	var keys []string
	for k := range c.Trace {
		keys = append(keys, k)
	}

	return keys
}

func (c *GenericTelemetryCarrier) Set(key, val string) {
	c.Trace[key] = val
}

// Use if one of these cases apply:
// - Moving context between different services or processes.
// - Merging or transplanting the trace context from one operation to another (for example, when handling chained requests).
// - Passing context from one part of an application to another that has its own context, but you want to carry over the telemetry information from the original context.
func TransplantTelemetryInformation(a, b context.Context) context.Context {
	carrier := &GenericTelemetryCarrier{
		Trace: make(map[string]string),
	}
	prop := otel.GetTextMapPropagator()
	prop.Inject(a, carrier)
	ctx := prop.Extract(b, carrier)

	return ctx
}
