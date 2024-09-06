package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
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

// InjectTraceParent injects the given traceparent into a new context and returns it with the parent span.
// The tracer is automatically obtained from the global OpenTelemetry TracerProvider.
func InjectTraceParent(ctx context.Context, traceParent string, traceName string) (context.Context, trace.Span, error) {
	// Set up the propagation map with the traceparent
	carrier := propagation.MapCarrier{
		"traceparent": traceParent,
	}

	propagator := propagation.TraceContext{}
	newCtx := propagator.Extract(ctx, carrier)

	tracer := otel.GetTracerProvider().Tracer(instrumentationName)

	// Start a new span with this context, making it the parent span
	newCtx, span := tracer.Start(newCtx, traceName)
	if span.SpanContext().IsValid() {
		return newCtx, span, nil
	}

	return newCtx, span, fmt.Errorf("failed to inject traceparent as parent span")
}

// Span starts a new span with the provided name as a child of the context with tracing.
// It returns a function that ends the span when called.
func Span(ctxWithTracing context.Context, name string) (context.Context, func(), error) {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	ctx, span := tracer.Start(ctxWithTracing, name)
	if !span.SpanContext().IsValid() {
		return nil, nil, fmt.Errorf("failed to start span for %s", name)
	}

	endSpan := func() {
		defer span.End()
	}

	return ctx, endSpan, nil
}
