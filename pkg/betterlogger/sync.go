package betterlogger

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

// SyncAttributes holds metadata specific to a Sync component, helpful for logging.
type SyncAttributes struct {
	SyncID    string // Unique identifier for the Sync
	Namespace string // Namespace of the Sync
}

// LogSyncDebug logs a debug message with Sync attributes.
func LogSyncDebug(ctx context.Context, attr SyncAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("activity", attr.SyncID),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithTraceAndAttributes(ctx, slog.LevelDebug, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogSyncInfo logs an info message with Sync attributes.
func LogSyncInfo(ctx context.Context, attr SyncAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("activity", attr.SyncID),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithTraceAndAttributes(ctx, slog.LevelInfo, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogSyncWarn logs a warning message with Sync attributes.
func LogSyncWarn(ctx context.Context, attr SyncAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("activity", attr.SyncID),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithTraceAndAttributes(ctx, slog.LevelWarn, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogSyncError logs an error message with Sync attributes.
func LogSyncError(ctx context.Context, attr SyncAttributes, msg string, err error, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("activity", attr.SyncID),
		slog.String("namespace", attr.Namespace),
		slog.String("error", err.Error()),
	}
	internal.LogWithTraceAndAttributes(ctx, slog.LevelError, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}
