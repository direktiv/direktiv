package tracing_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/stretchr/testify/assert"
)

// Helper to initialize and return the shutdown function for telemetry
func initTestWithTelemetry() func() {
	shutdown, err := tracing.InitTelemetry(context.Background(), "", "test-service", "test-instrumentation")
	if err != nil {
		panic("Failed to initialize test telemetry: " + err.Error())
	}
	return shutdown
}

// Test case for injecting traceparent using mock telemetry
func TestInjectWithMockTraceParent(t *testing.T) {
	shutdown := initTestWithTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Valid traceparent
	traceParent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	newCtx, span, err := tracing.InjectTraceParent(ctx, traceParent, "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.True(t, span.SpanContext().IsValid())

	// Invalid traceparent format
	traceParent = "invalid-traceparent"
	newCtx, span, err = tracing.InjectTraceParent(ctx, traceParent, "test-span")
	assert.NoError(t, err) // always needs to return a valid span
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)

	// Empty traceparent
	traceParent = ""
	newCtx, span, err = tracing.InjectTraceParent(ctx, traceParent, "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
}

// Test case for creating spans with mock telemetry
func TestNewSpanWithMock(t *testing.T) {
	shutdown := initTestWithTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Valid span creation
	newCtx, endSpan, err := tracing.NewSpan(ctx, "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, endSpan)
	endSpan()

	// Ensure span is properly ended
	// Trying to create another span and checking that previous context is still valid
	_, endSpan2, err := tracing.NewSpan(ctx, "test-span-2")
	assert.NoError(t, err)
	assert.NotNil(t, endSpan2)
	endSpan2()
}

// Test case for span attributes and multiple spans
func TestSpanAttributesAndMultipleSpans(t *testing.T) {
	shutdown := initTestWithTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Create first span
	ctx1, endSpan1, err := tracing.NewSpan(ctx, "span-1")
	assert.NoError(t, err)
	assert.NotNil(t, ctx1)
	assert.NotNil(t, endSpan1)

	// Create second span within the first span's context
	ctx2, endSpan2, err := tracing.NewSpan(ctx1, "span-2")
	assert.NoError(t, err)
	assert.NotNil(t, ctx2)
	assert.NotNil(t, endSpan2)

	// Ensure both spans can be ended properly
	endSpan2()
	endSpan1()
}

// Test case to ensure spans are not interfering
func TestSpanIsolation(t *testing.T) {
	shutdown := initTestWithTelemetry()
	defer shutdown()

	ctx := context.Background()

	// Create two independent spans
	ctx1, endSpan1, err := tracing.NewSpan(ctx, "isolated-span-1")
	assert.NoError(t, err)
	assert.NotNil(t, ctx1)
	assert.NotNil(t, endSpan1)

	ctx2, endSpan2, err := tracing.NewSpan(ctx, "isolated-span-2")
	assert.NoError(t, err)
	assert.NotNil(t, ctx2)
	assert.NotNil(t, endSpan2)

	// Ensure both spans are independent
	assert.NotEqual(t, ctx1, ctx2)

	// End spans
	endSpan1()
	endSpan2()
}

func TestSpanAttributePropagation(t *testing.T) {
	shutdown := initTestWithTelemetry()
	defer shutdown()

	ctx := context.WithValue(context.Background(), tracing.NamespaceKey, "test-namespace")

	ctx1, endSpan1, err := tracing.NewSpan(ctx, "parent-span")
	assert.NoError(t, err)
	assert.NotNil(t, ctx1)
	assert.NotNil(t, endSpan1)

	ctx2, endSpan2, err := tracing.NewSpan(ctx1, "child-span")
	assert.NoError(t, err)
	assert.NotNil(t, ctx2)
	assert.NotNil(t, endSpan2)

	// Ensure attribute exists in the new span's context
	assert.Equal(t, "test-namespace", ctx2.Value(tracing.NamespaceKey))

	endSpan2()
	endSpan1()
}
