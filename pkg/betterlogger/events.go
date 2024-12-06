package betterlogger

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

// CloudEventBusAttributes holds metadata specific to the cloud event bus.
type CloudEventBusAttributes struct {
	EventID   string // Unique identifier for the event
	Source    string // Source of the event
	Subject   string // Subject of the event
	EventType string // Type of the event
	Namespace string // Namespace of the event bus
}

// LogCloudEventDebug logs a debug message with cloud event bus attributes.
func LogCloudEventDebug(ctx context.Context, attr CloudEventBusAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("event_id", attr.EventID),
		slog.String("source", attr.Source),
		slog.String("subject", attr.Subject),
		slog.String("event_type", attr.EventType),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelDebug, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogCloudEventInfo logs an info message with cloud event bus attributes.
func LogCloudEventInfo(ctx context.Context, attr CloudEventBusAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("event_id", attr.EventID),
		slog.String("source", attr.Source),
		slog.String("subject", attr.Subject),
		slog.String("event_type", attr.EventType),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelInfo, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogCloudEventWarn logs a warning message with cloud event bus attributes.
func LogCloudEventWarn(ctx context.Context, attr CloudEventBusAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("event_id", attr.EventID),
		slog.String("source", attr.Source),
		slog.String("subject", attr.Subject),
		slog.String("event_type", attr.EventType),
		slog.String("namespace", attr.Namespace),
	}
	internal.LogWithAttributes(ctx, slog.LevelWarn, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogCloudEventError logs an error message with cloud event bus attributes.
func LogCloudEventError(ctx context.Context, attr CloudEventBusAttributes, msg string, err error, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("event_id", attr.EventID),
		slog.String("source", attr.Source),
		slog.String("subject", attr.Subject),
		slog.String("event_type", attr.EventType),
		slog.String("namespace", attr.Namespace),
		slog.String("error", err.Error()),
	}
	internal.LogWithAttributes(ctx, slog.LevelError, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}
