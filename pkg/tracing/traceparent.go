package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// ExtractTraceParent extracts the traceparent from the context.
func ExtractTraceParent(ctx context.Context) (string, error) {
	// Use the TraceContext propagator to extract traceparent
	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)

	traceParent := carrier.Get("traceparent")
	if traceParent == "" {
		return "", fmt.Errorf("no traceparent found in the context")
	}

	return traceParent, nil
}

func TraceParentToTraceID(ctx context.Context, traceParent string) (string, error) {
	// Set up the propagation map with the traceparent
	carrier := propagation.MapCarrier{
		"traceparent": traceParent,
	}

	propagator := propagation.TraceContext{}
	newCtx := propagator.Extract(ctx, carrier)
	span := trace.SpanFromContext(newCtx)
	if span.SpanContext().TraceID().IsValid() {
		return span.SpanContext().TraceID().String(), nil
	}

	return "", fmt.Errorf("failed extract trace-id from traceparent")
}

// InjectTraceParent injects the given traceparent into a new context and returns it with the parent span.
// The tracer is automatically obtained from the global OpenTelemetry TracerProvider.
// The function ensures that tracing always happens, even if the provided traceparent is invalid.
//
// This function operates in two modes:
// 1. **Valid Traceparent** → A properly linked span is created, attributes are set, and no error is returned.
// 2. **Invalid Traceparent** → A span is still created (ensuring traceability), attributes are set, but an error is returned.
func InjectTraceParent(ctx context.Context, traceParent string, traceName string) (context.Context, trace.Span, error) {
	// Set up the propagation map with the traceparent
	carrier := propagation.MapCarrier{
		"traceparent": traceParent,
	}

	propagator := propagation.TraceContext{}
	newCtx := propagator.Extract(ctx, carrier)

	tracer := otel.GetTracerProvider().Tracer(instrumentationName)

	// Start a new span with this context, making it the parent span
	// # Context Handling Explanation:
	// - `ctx` is the original context, which may contain useful metadata (namespace, instance, workflow, etc.).
	// - `newCtx` is derived by extracting the `traceparent` from `ctx` and is used as the parent for the new span.
	// - We use `ctx` (not `newCtx`) to retrieve attributes because `ctx` may have pre-existing metadata before tracing was injected.
	newCtx, span := tracer.Start(newCtx, traceName)
	if span.SpanContext().IsValid() {
		// Retrieve attributes from the original context (`ctx`), not `newCtx`
		// to ensure we capture metadata that existed before tracing was injected.
		attr := GetCoreAttributes(ctx)
		kv := make([]attribute.KeyValue, 0, len(attr)*2)
		for k, v := range attr {
			kv = append(kv, attribute.String(k, fmt.Sprint(v)))
		}
		span.SetAttributes(kv...)

		return newCtx, span, nil
	}

	// Even if the traceparent was invalid, we still create a span and attach attributes.
	attr := GetCoreAttributes(ctx)
	kv := make([]attribute.KeyValue, 0, len(attr)*2)
	for k, v := range attr {
		kv = append(kv, attribute.String(k, fmt.Sprint(v)))
	}
	span.SetAttributes(kv...)

	return newCtx, span, fmt.Errorf("failed to inject traceparent as parent span")
}

// NewSpan starts a new span with the provided name as a child of the context with tracing.
// It returns a function that ends the span when called.
func NewSpan(ctx context.Context, name string) (context.Context, func(), error) {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	ctx2, span := tracer.Start(ctx, name)
	if !span.SpanContext().IsValid() {
		return ctx, func() {}, fmt.Errorf("failed to start span for %s", name)
	}
	ctx = ctx2
	attr := GetCoreAttributes(ctx)
	kv := make([]attribute.KeyValue, 0, len(attr)*2)
	for k, v := range attr {
		kv = append(kv, attribute.String(k, fmt.Sprint(v)))
	}
	endSpan := func() {
		span.End()
	}
	span.SetAttributes(kv...)

	return ctx, endSpan, nil
}
