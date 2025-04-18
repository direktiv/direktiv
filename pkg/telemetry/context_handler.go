package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
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
	l := ctx.Value(logObjectCtx)

	res := make([]slog.Attr, 0)
	res = append(res, slog.Attr{
		Key:   "nanos",
		Value: slog.AnyValue(time.Now().UTC().UnixNano()),
	})

	if l != nil {
		// double marshal
		b, err := json.Marshal(l)
		if err != nil {
			slog.Error("can not marshal context", slog.Any("error", err))
			return h.innerHandler.Handle(ctx, rec)
		}

		var attrs map[string]interface{}
		err = json.Unmarshal(b, &attrs)
		if err != nil {
			slog.Error("can not unmarshal context", slog.Any("error", err))
			return h.innerHandler.Handle(ctx, rec)
		}

		for k, v := range attrs {
			res = append(res, slog.Attr{
				Key:   strings.ToLower(k),
				Value: slog.AnyValue(fmt.Sprintf("%v", v)),
			})
		}

		return h.innerHandler.WithAttrs(res).Handle(ctx, rec)
	}

	return h.innerHandler.WithAttrs(res).Handle(ctx, rec)
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
