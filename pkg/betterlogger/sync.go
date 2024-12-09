package betterlogger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

// buildSyncAttributes constructs the base attributes for mirror logging.
func buildSyncAttributes(attr SyncAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("activity", attr.SyncID),
		slog.String("namespace", attr.Namespace),
		slog.String("track", fmt.Sprintf("%v.%v", "activity", attr.SyncID)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// SyncAttributes holds metadata specific to a Sync component, helpful for logging.
type SyncAttributes struct {
	SyncID    string // Unique identifier for the Sync
	Namespace string // Namespace of the Sync
}

// LogSyncDebug logs a debug message with Sync attributes.
func LogSyncDebug(ctx context.Context, attr SyncAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelDebug, msg, buildSyncAttributes(attr, sAttr...)...)
}

// LogSyncInfo logs an info message with Sync attributes.
func LogSyncInfo(ctx context.Context, attr SyncAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelInfo, msg, buildSyncAttributes(attr, sAttr...)...)
}

// LogSyncWarn logs a warning message with Sync attributes.
func LogSyncWarn(ctx context.Context, attr SyncAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelWarn, msg, buildSyncAttributes(attr, sAttr...)...)
}

// LogSyncError logs an error message with Sync attributes.
func LogSyncError(ctx context.Context, attr SyncAttributes, msg string, err error, sAttr ...slog.Attr) {
	errorAttr := slog.String("error", err.Error())
	internal.LogWithTraceAndAttributes(ctx, slog.LevelError, msg, buildSyncAttributes(attr, append(sAttr, errorAttr)...)...)
}
