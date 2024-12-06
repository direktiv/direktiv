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

// LogGatewayDebug logs a debug message with gateway attributes.
func LogGatewayDebug(ctx context.Context, attr GatewayAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("component", "gateway"),
		slog.String("gateway_plugin", attr.Plugin),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelDebug, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogGatewayInfo logs an info message with gateway attributes.
func LogGatewayInfo(ctx context.Context, attr GatewayAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("component", "gateway"),
		slog.String("gateway_plugin", attr.Plugin),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelInfo, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogGatewayWarn logs a warning message with gateway attributes.
func LogGatewayWarn(ctx context.Context, attr GatewayAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("component", "gateway"),
		slog.String("gateway_plugin", attr.Plugin),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelWarn, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogGatewayError logs an error message with gateway attributes.
func LogGatewayError(ctx context.Context, attr GatewayAttributes, msg string, err error, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("component", "gateway"),
		slog.String("gateway_plugin", attr.Plugin),
		slog.String("namespace", attr.Namespace),
		slog.String("error", err.Error()),
	}
	internal.LogWithAttributes(ctx, slog.LevelError, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}
