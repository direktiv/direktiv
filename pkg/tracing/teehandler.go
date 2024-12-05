package tracing

import (
	"context"
	"fmt"
	"log/slog"
)

var _ slog.Handler = TeeHandler{}

type TeeHandler []slog.Handler

// Enabled implements slog.Handler.
func (s TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return s[0].Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (s TeeHandler) Handle(ctx context.Context, rec slog.Record) error {
	var errs error
	for _, logger := range s {
		err := logger.Handle(ctx, rec)
		if err != nil {
			errs = fmt.Errorf("%w :%w", errs, err)
		}
	}

	return errs
}

// WithAttrs implements slog.Handler.
func (s TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newChain := make(TeeHandler, len(s))
	for i, logger := range s {
		newChain[i] = logger.WithAttrs(attrs)
	}

	return newChain
}

// WithGroup implements slog.Handler.
func (s TeeHandler) WithGroup(name string) slog.Handler {
	newChain := make(TeeHandler, len(s))
	for i, logger := range s {
		newChain[i] = logger.WithGroup(name)
	}

	return newChain
}
