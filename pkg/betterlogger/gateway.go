package betterlogger

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

// GatewayAttributes holds metadata specific to a gateway component, helpful for logging.
type GatewayAttributes struct {
	Plugin    string // Name for the gateway-plugin
	Namespace string // Namespace of the gateway
}

// buildGatewayAttributes constructs the base attributes for gateway logging.
func buildGatewayAttributes(attr GatewayAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("component", "gateway"),
		slog.String("gateway_plugin", attr.Plugin),
		slog.String("namespace", attr.Namespace),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// LogGatewayDebug logs a debug message with gateway attributes.
func LogGatewayDebug(ctx context.Context, attr GatewayAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelDebug, msg, buildGatewayAttributes(attr, sAttr...)...)
}

// LogGatewayInfo logs an info message with gateway attributes.
func LogGatewayInfo(ctx context.Context, attr GatewayAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelInfo, msg, buildGatewayAttributes(attr, sAttr...)...)
}

// LogGatewayWarn logs a warning message with gateway attributes.
func LogGatewayWarn(ctx context.Context, attr GatewayAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelWarn, msg, buildGatewayAttributes(attr, sAttr...)...)
}

// LogGatewayError logs an error message with gateway attributes.
func LogGatewayError(ctx context.Context, attr GatewayAttributes, msg string, err error, sAttr ...slog.Attr) {
	errorAttr := slog.String("error", err.Error())
	internal.LogWithTraceAndAttributes(ctx, slog.LevelError, msg, buildGatewayAttributes(attr, append(sAttr, errorAttr)...)...)
}
