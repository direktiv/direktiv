package tracing_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Helper to initialize and return the shutdown function for no telemetry
func initTestWithNoTelemetry() func() {
	shutdown, err := tracing.InitTelemetry(context.Background(), "", "test-service", "test-instrumentation")
	if err != nil {
		panic("Failed to initialize test telemetry: " + err.Error())
	}
	return shutdown
}

// Helper to initialize and return the shutdown function for mock telemetry
func initTestWithMockTelemetry() func() {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
	return func() {
		_ = tp.Shutdown(context.Background())
	}
}

// Test case for injecting a traceparent using real telemetry
func TestInjectTraceParent(t *testing.T) {
	shutdown := initTestWithNoTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Valid traceparent
	traceParent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	newCtx, span, err := tracing.InjectTraceParent(ctx, traceParent, "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.True(t, span.SpanContext().IsValid())

	// Invalid traceparent
	traceParent = "invalid-traceparent"
	newCtx, span, err = tracing.InjectTraceParent(ctx, traceParent, "test-span")
	assert.Error(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.False(t, span.SpanContext().IsValid())
}

// Test case for creating spans with real telemetry
func TestNewSpan(t *testing.T) {
	shutdown := initTestWithNoTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Valid span creation
	newCtx, endSpan, err := tracing.NewSpan(ctx, "test-span")
	assert.Error(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, endSpan)
	endSpan()

	// Invalid span creation with an empty name
	invalidCtx := context.TODO()
	newCtx, endSpan, err = tracing.NewSpan(invalidCtx, "")
	assert.Error(t, err)
	assert.NotNil(t, newCtx)
	assert.EqualError(t, err, "failed to start span for ")
}

// Test case for injecting traceparent using mock telemetry
func TestInjectWithMockTraceParent(t *testing.T) {
	shutdown := initTestWithMockTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Valid traceparent
	traceParent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	newCtx, span, err := tracing.InjectTraceParent(ctx, traceParent, "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.True(t, span.SpanContext().IsValid())

	// Invalid traceparent
	traceParent = "invalid-traceparent"
	newCtx, span, err = tracing.InjectTraceParent(context.Background(), traceParent, "test-span")
	assert.NoError(t, err) // always need to valid span no matter what
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
}

// Test case for creating spans with mock telemetry
func TestNewSpanWithMock(t *testing.T) {
	shutdown := initTestWithMockTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Valid span creation
	newCtx, endSpan, err := tracing.NewSpan(ctx, "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, endSpan)
	endSpan()
}
