package betterlogger

import (
	"context"
	"fmt"
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

// buildCloudEventAttributes constructs the base attributes for cloud event logging.
func buildCloudEventAttributes(attr CloudEventBusAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("event_id", attr.EventID),
		slog.String("source", attr.Source),
		slog.String("subject", attr.Subject),
		slog.String("event_type", attr.EventType),
		slog.String("namespace", attr.Namespace),
		slog.String("track", fmt.Sprintf("%v.%v", "namespace", attr.Namespace)),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// LogCloudEventDebug logs a debug message with cloud event bus attributes.
func LogCloudEventDebug(ctx context.Context, attr CloudEventBusAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelDebug, msg, buildCloudEventAttributes(attr, sAttr...)...)
}

// LogCloudEventInfo logs an info message with cloud event bus attributes.
func LogCloudEventInfo(ctx context.Context, attr CloudEventBusAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelInfo, msg, buildCloudEventAttributes(attr, sAttr...)...)
}

// LogCloudEventWarn logs a warning message with cloud event bus attributes.
func LogCloudEventWarn(ctx context.Context, attr CloudEventBusAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelWarn, msg, buildCloudEventAttributes(attr, sAttr...)...)
}

// LogCloudEventError logs an error message with cloud event bus attributes.
func LogCloudEventError(ctx context.Context, attr CloudEventBusAttributes, msg string, err error, sAttr ...slog.Attr) {
	errorAttr := slog.String("error", err.Error())
	internal.LogWithTraceAndAttributes(ctx, slog.LevelError, msg, buildCloudEventAttributes(attr, append(sAttr, errorAttr)...)...)
}
