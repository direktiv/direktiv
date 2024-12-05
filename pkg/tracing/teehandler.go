package tracing

import (
	"context"
	"fmt"
	"log/slog"
)

var _ slog.Handler = TeeHandler{}

// TeeHandler allows log records to be sent to multiple handlers simultaneously. It implements the `slog.Handler`
// interface, forwarding each log entry to all registered handlers.
//
// This handler is useful when you want logs to be processed by multiple targets (e.g., logging to stdout and sending
// logs to a remote server) without duplicating the log calls.
//
// By chaining multiple handlers together, it ensures that the same log record can be written to different destinations.
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
