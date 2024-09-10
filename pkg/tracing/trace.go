package tracing

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/cloudevents/sdk-go/v2/event"
// 	"github.com/direktiv/direktiv/pkg/flow/nohome/recipient"
// 	"github.com/direktiv/direktiv/pkg/utils"
// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/attribute"
// 	"go.opentelemetry.io/otel/codes"
// 	"go.opentelemetry.io/otel/trace"
// )

// // Carrier controls the trace.
// type Carrier struct {
// 	Trace map[string]string
// }

// // Get returns a value of a trace.
// func (c *Carrier) Get(key string) string {
// 	v := c.Trace[key]
// 	return v
// }

// // Keys returns all the keys of the trace.
// func (c *Carrier) Keys() []string {
// 	var keys []string
// 	for k := range c.Trace {
// 		keys = append(keys, k)
// 	}

// 	return keys
// }

// // Set sets a trace key and value.
// func (c *Carrier) Set(key, val string) {
// 	c.Trace[key] = val
// }

// func dbTrace(ctx context.Context) *Carrier {
// 	carrier := &Carrier{
// 		Trace: make(map[string]string),
// 	}
// 	prop := otel.GetTextMapPropagator()
// 	prop.Inject(ctx, carrier)

// 	return carrier
// }

// func traceAddWorkflowInstance(ctx context.Context, im *instanceMemory) {
// 	span := trace.SpanFromContext(ctx)

// 	m := im.instance.GetAttributes(recipient.Instance)
// 	delete(m, "recipientType")

// 	attrs := make([]attribute.KeyValue, 0)
// 	for k, v := range m {
// 		attrs = append(attrs, attribute.KeyValue{
// 			Key:   attribute.Key(k),
// 			Value: attribute.StringValue(v),
// 		})
// 	}

// 	span.SetAttributes(attrs...)
// }

// func traceFullAddWorkflowInstance(ctx context.Context, im *instanceMemory) (context.Context, error) {
// 	traceAddWorkflowInstance(ctx, im)
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	ctx, span := tr.Start(ctx, "new-workflow-instance", trace.WithSpanKind(trace.SpanKindInternal))
// 	defer span.End()
// 	traceAddWorkflowInstance(ctx, im)

// 	x := dbTrace(ctx)
// 	s := utils.Marshal(x)

// 	im.instance.TelemetryInfo.TraceID = s
// 	data, err := im.instance.TelemetryInfo.MarshalJSON()
// 	if err != nil {
// 		return nil, err
// 	}

// 	im.updateArgs.TelemetryInfo = &data

// 	return ctx, nil
// }

// func traceStateError(ctx context.Context, err error) {
// 	span := trace.SpanFromContext(ctx)
// 	span.SetStatus(codes.Error, err.Error())
// 	span.AddEvent(fmt.Sprintf("state error: %v", err.Error()))
// }

// func traceSubflowInvoke(ctx context.Context, name, child string) {
// 	span := trace.SpanFromContext(ctx)

// 	span.SetAttributes(
// 		attribute.KeyValue{
// 			Key:   "child-instance",
// 			Value: attribute.StringValue(child),
// 		},
// 	)

// 	span.AddEvent(fmt.Sprintf("Calling subflow: %s (%s)", name, child))
// }

// func traceStateGenericBegin(ctx context.Context, im *instanceMemory) (context.Context, func(), error) {
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	prop := otel.GetTextMapPropagator()
// 	var span trace.Span

// 	carrier := new(Carrier)

// 	err := json.Unmarshal([]byte(im.instance.TelemetryInfo.TraceID), carrier)
// 	if err != nil {
// 		return ctx, nil, err
// 	}

// 	ctx = prop.Extract(ctx, carrier)

// 	ctx, span = tr.Start(ctx, im.logic.GetType().String(), trace.WithSpanKind(trace.SpanKindInternal))

// 	x := dbTrace(ctx)
// 	s := utils.Marshal(x)

// 	im.instance.TelemetryInfo.SpanID = s
// 	data, err := im.instance.TelemetryInfo.MarshalJSON()
// 	if err != nil {
// 		return ctx, nil, err
// 	}

// 	im.updateArgs.TelemetryInfo = &data

// 	finish := func() {
// 		span.End()
// 	}

// 	return ctx, finish, nil
// }

// func traceStateGenericLogicThread(ctx context.Context, im *instanceMemory) (context.Context, func(), error) {
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	prop := otel.GetTextMapPropagator()
// 	var span trace.Span

// 	carrier := new(Carrier)
// 	err := json.Unmarshal([]byte(im.instance.TelemetryInfo.SpanID), carrier)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	ctx = prop.Extract(ctx, carrier)

// 	ctx, span = tr.Start(ctx, fmt.Sprintf("%s-logic", im.logic.GetType().String()), trace.WithSpanKind(trace.SpanKindInternal))

// 	finish := func() {
// 		span.End()
// 	}

// 	return ctx, finish, nil
// }

// func traceActionResult(ctx context.Context, results *actionResultPayload) {
// 	span := trace.SpanFromContext(ctx)
// 	span.SetAttributes(
// 		attribute.KeyValue{
// 			Key:   "returned-action",
// 			Value: attribute.StringValue(results.ActionID),
// 		},
// 	)
// }

// func traceAddtoEventlog(ctx context.Context) (context.Context, func()) {
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	ctx, span := tr.Start(ctx, "addToEventLog", trace.WithSpanKind(trace.SpanKindInternal))
// 	finish := func() {
// 		span.End()
// 	}

// 	return ctx, finish
// }

// func traceBrokerMessage(ctx context.Context, ev event.Event) (context.Context, func()) {
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	ctx, span := tr.Start(ctx, "BroadcastCloudevent", trace.WithSpanKind(trace.SpanKindInternal))
// 	span.SetAttributes(
// 		attribute.KeyValue{
// 			Key:   "event-registered",
// 			Value: attribute.StringValue(ev.Source() + "-" + ev.ID()),
// 		},
// 	)
// 	finish := func() {
// 		span.End()
// 	}

// 	return ctx, finish
// }

// func traceGetListenersByTopic(ctx context.Context, topic string) (context.Context, func()) {
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	ctx, span := tr.Start(ctx, "GetListenersByTopic", trace.WithSpanKind(trace.SpanKindInternal))
// 	span.SetAttributes(
// 		attribute.KeyValue{
// 			Key:   "topic",
// 			Value: attribute.StringValue(topic),
// 		},
// 	)
// 	finish := func() {
// 		span.End()
// 	}

// 	return ctx, finish
// }

// func traceProcessingMessage(ctx context.Context) (context.Context, func()) {
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	ctx, span := tr.Start(ctx, "processingCloudevent", trace.WithSpanKind(trace.SpanKindInternal))

// 	finish := func() {
// 		span.End()
// 	}

// 	return ctx, finish
// }

// func traceMessageTrigger(ctx context.Context, triggerDescription string) (context.Context, func()) { //nolint:unparam
// 	tp := otel.GetTracerProvider()
// 	tr := tp.Tracer("direktiv/flow")
// 	ctx, span := tr.Start(ctx, "triggered-by-event", trace.WithSpanKind(trace.SpanKindInternal))
// 	span.SetAttributes(
// 		attribute.KeyValue{
// 			Key:   "trigger-desc",
// 			Value: attribute.StringValue(triggerDescription),
// 		},
// 	)
// 	finish := func() {
// 		span.End()
// 	}

// 	return ctx, finish
// }

// Trace adds an event to the current span from the context.
func Trace(ctx context.Context, msg string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(msg)
	} else {
		slog.Warn("No active span found in context for tracing event.", "event", msg)
	}
}

func TraceGWHTTPRequest(ctx context.Context, r *http.Request, instrumentationName string) (cleanup func()) {
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
