package tracing

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

var _ slog.Handler = &EventHandler{}

type EventHandler struct{}

// Enabled implements slog.Handler.
func (e EventHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return slog.Default().Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (e EventHandler) Handle(ctx context.Context, rec slog.Record) error {
	addEvent(ctx, rec.Message)

	return nil
}

// WithAttrs implements slog.Handler.
func (e EventHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &EventHandler{}
}

// WithGroup implements slog.Handler.
func (e EventHandler) WithGroup(name string) slog.Handler {
	return &EventHandler{}
}

func addEvent(ctx context.Context, msg string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(msg)
	}
}
