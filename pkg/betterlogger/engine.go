package betterlogger

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/betterlogger/internal"
)

// InstanceAttributes holds common metadata for an instance, which is helpful for logging.
type InstanceAttributes struct {
	Namespace    string // Namespace where the instance belongs
	InstanceID   string // Unique identifier for the instance
	WorkflowPath string // Path of the workflow the instance belongs to
}

// InstanceMemoryAttributes holds common metadata for an instance, which is helpful for logging.
type InstanceMemoryAttributes struct {
	InstanceAttributes
	State string // Memory state of the instance
}

// buildInstanceAttributes constructs the base attributes for instance logging.
func buildInstanceAttributes(attr InstanceAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := []slog.Attr{
		slog.String("namespace", attr.Namespace),
		slog.String("instance", attr.InstanceID),
		slog.String("workflow", attr.WorkflowPath),
	}

	return internal.MergeAttributes(baseAttrs, additionalAttrs...)
}

// buildInstanceMemoryAttributes constructs the base attributes for instance memory logging.
func buildInstanceMemoryAttributes(attr InstanceMemoryAttributes, additionalAttrs ...slog.Attr) []slog.Attr {
	baseAttrs := buildInstanceAttributes(attr.InstanceAttributes)
	memoryAttr := slog.String("state", attr.State)

	return internal.MergeAttributes(baseAttrs, append(additionalAttrs, memoryAttr)...)
}

// LogInstanceDebug logs a debug message with instance attributes.
func LogInstanceDebug(ctx context.Context, attr InstanceAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelDebug, msg, buildInstanceAttributes(attr, sAttr...)...)
}

// LogInstanceInfo logs an info message with instance attributes.
func LogInstanceInfo(ctx context.Context, attr InstanceAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelInfo, msg, buildInstanceAttributes(attr, sAttr...)...)
}

// LogInstanceWarn logs a warning message with instance attributes.
func LogInstanceWarn(ctx context.Context, attr InstanceAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelWarn, msg, buildInstanceAttributes(attr, sAttr...)...)
}

// LogInstanceError logs an error message with instance attributes.
func LogInstanceError(ctx context.Context, attr InstanceAttributes, msg string, err error, sAttr ...slog.Attr) {
	errorAttr := slog.String("error", err.Error())
	internal.LogWithTraceAndAttributes(ctx, slog.LevelError, msg, buildInstanceAttributes(attr, append(sAttr, errorAttr)...)...)
}

// LogInstanceMemoryDebug logs a debug message with instance memory attributes.
func LogInstanceMemoryDebug(ctx context.Context, attr InstanceMemoryAttributes, msg string, sAttr ...slog.Attr) {
	internal.LogWithTraceAndAttributes(ctx, slog.LevelDebug, msg, buildInstanceMemoryAttributes(attr, sAttr...)...)
}
