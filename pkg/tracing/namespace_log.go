package tracing

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// NewNamespaceLogger creates a logger with namespace and track information.

func NewNamespaceLogger(namespaceName string) slog.Logger {
	return *slog.With("namespace", namespaceName, "track", BuildNamespaceTrack(namespaceName))
}

// NewNamespaceLoggerWithTrace creates a logger with namespace, track, and tracing information.

func NewNamespaceLoggerWithTrace(ctx context.Context, namespaceName string) slog.Logger {
	logger := NewNamespaceLogger(namespaceName)
	span := trace.SpanFromContext(ctx)
	span.SpanContext()

	return *logger.With("trace", span.SpanContext().TraceID().String(), "span", span.SpanContext().SpanID().String())

}
