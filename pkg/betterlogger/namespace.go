package betterlogger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

// buildGatewayAttributes constructs the base attributes for gateway logging.
func buildNamespaceAttributes(namespace string, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("namespace", namespace),
		slog.String("track", fmt.Sprintf("%v.%v", "namespace", namespace)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// Namespace logging helpers

// LogNamespaceDebug logs a debug message scoped to a namespace.
func LogNamespaceDebug(ctx context.Context, namespace string, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelDebug, msg, buildNamespaceAttributes(namespace, sAttr...)...)
}

// LogNamespaceInfo logs an info message scoped to a namespace.
func LogNamespaceInfo(ctx context.Context, namespace string, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelInfo, msg, buildNamespaceAttributes(namespace, sAttr...)...)
}

// LogNamespaceWarn logs a warning message scoped to a namespace.
func LogNamespaceWarn(ctx context.Context, namespace string, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelWarn, msg, buildNamespaceAttributes(namespace, sAttr...)...)
}

// LogNamespaceError logs an error message scoped to a namespace.
func LogNamespaceError(ctx context.Context, namespace string, msg string, err error, sAttr ...slog.Attr) {
	errorAttr := slog.String("error", err.Error())
	internal.LogWithTraceAndAttributes(ctx, slog.LevelError, msg, buildNamespaceAttributes(namespace, append(sAttr, errorAttr)...)...)
}
