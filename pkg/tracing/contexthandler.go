package tracing

import (
	"context"
	"log/slog"
)

// ContextHandler wraps a slog.Handler (e.g., JSON handler) and processes slogFields from the context.
type ContextHandler struct {
	innerHandler slog.Handler
}

func NewContextHandler(innerHandler slog.Handler) slog.Handler {
	return &ContextHandler{
		innerHandler: innerHandler,
	}
}

// Enabled implements slog.Handler.
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.innerHandler.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *ContextHandler) Handle(ctx context.Context, rec slog.Record) error {
	if attrs := GetAttributes(ctx); len(attrs) > 0 {
		res := make([]slog.Attr, 0, len(attrs)*2)
		for k, v := range attrs {
			res = append(res, slog.Attr{Key: k, Value: slog.AnyValue(v)})
		}

		return h.innerHandler.WithAttrs(res).Handle(ctx, rec)
	}

	// Pass the record to the inner handler
	return h.innerHandler.Handle(ctx, rec)
}

// WithAttrs implements slog.Handler.
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewContextHandler(h.innerHandler.WithAttrs(attrs))
}

// WithGroup implements slog.Handler.
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return NewContextHandler(h.innerHandler.WithGroup(name))
}

var _ slog.Handler = &ContextHandler{}
