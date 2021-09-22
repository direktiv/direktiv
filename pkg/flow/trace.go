package flow

import (
	"context"

	"go.opentelemetry.io/otel"
)

type Carrier struct {
	Trace map[string]string
}

func (c *Carrier) Get(key string) string {
	v, _ := c.Trace[key]
	return v
}

func (c *Carrier) Keys() []string {
	var keys []string
	for k := range c.Trace {
		keys = append(keys, k)
	}
	return keys
}

func (c *Carrier) Set(key, val string) {
	c.Trace[key] = val
}

func dbTrace(ctx context.Context) *Carrier {
	carrier := &Carrier{
		Trace: make(map[string]string),
	}
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, carrier)
	return carrier
}
