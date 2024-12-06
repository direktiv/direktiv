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

// LogInstanceDebug logs a debug message with instance attributes.
func LogInstanceDebug(ctx context.Context, attr InstanceAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", attr.Namespace),
		slog.String("instance", attr.InstanceID),
		slog.String("workflow", attr.WorkflowPath),
	}
	internal.LogWithAttributes(ctx, slog.LevelDebug, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogInstanceInfo logs an info message with instance attributes.
func LogInstanceInfo(ctx context.Context, attr InstanceAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", attr.Namespace),
		slog.String("instance", attr.InstanceID),
		slog.String("workflow", attr.WorkflowPath),
	}
	internal.LogWithAttributes(ctx, slog.LevelInfo, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogInstanceWarn logs a warning message with instance attributes.
func LogInstanceWarn(ctx context.Context, attr InstanceAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", attr.Namespace),
		slog.String("instance", attr.InstanceID),
		slog.String("workflow", attr.WorkflowPath),
	}
	internal.LogWithAttributes(ctx, slog.LevelWarn, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogInstanceError logs an error message with instance attributes.
func LogInstanceError(ctx context.Context, attr InstanceAttributes, msg string, err error, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", attr.Namespace),
		slog.String("instance", attr.InstanceID),
		slog.String("workflow", attr.WorkflowPath),
		slog.String("error", err.Error()),
	}
	internal.LogWithAttributes(ctx, slog.LevelError, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}

// LogInstanceMemoryDebug logs a debug message with instance memory attributes.
func LogInstanceMemoryDebug(ctx context.Context, attr InstanceMemoryAttributes, msg string, sAttr ...slog.Attr) {
	baseAttrs := []slog.Attr{
		slog.String("namespace", attr.Namespace),
		slog.String("instance", attr.InstanceID),
		slog.String("workflow", attr.WorkflowPath),
		slog.String("state", attr.State),
	}
	internal.LogWithAttributes(ctx, slog.LevelDebug, msg, internal.MergeAttributes(baseAttrs, sAttr...)...)
}
