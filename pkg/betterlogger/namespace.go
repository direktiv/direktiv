package betterlogger

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

// Namespace logging helpers

// LogNamespaceDebug logs a debug message scoped to a namespace.
func LogNamespaceDebug(ctx context.Context, namespace string, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelDebug, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogNamespaceInfo logs an info message scoped to a namespace.
func LogNamespaceInfo(ctx context.Context, namespace string, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelInfo, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogNamespaceWarn logs a warning message scoped to a namespace.
func LogNamespaceWarn(ctx context.Context, namespace string, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelWarn, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogNamespaceError logs an error message scoped to a namespace.
func LogNamespaceError(ctx context.Context, namespace string, msg string, err error, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", namespace),
		slog.String("error", err.Error()),
	}
	internal.LogWithAttributes(ctx, slog.LevelError, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}
