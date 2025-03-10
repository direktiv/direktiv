package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
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
	instance := ctx.Value(DirektivInstance)

	// only handle if there is something in the context
	if instance != nil {
		structVal := reflect.ValueOf(instance)
		fieldNum := structVal.NumField()
		structType := reflect.TypeOf(instance)

		res := make([]slog.Attr, 0)

		for i := range fieldNum {
			field := structVal.Field(i)
			fieldName := structType.Field(i).Name

			res = append(res, slog.Attr{Key: strings.ToLower(fieldName),
				Value: slog.AnyValue(fmt.Sprintf("%v", field))})
		}

		return h.innerHandler.WithAttrs(res).Handle(ctx, rec)
	}

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
