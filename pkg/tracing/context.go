package tracing

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
	"go.opentelemetry.io/otel/trace"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (level LogLevel) String() string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "DEBUG"
	}
}

func AddTag(ctx context.Context, key, value interface{}) context.Context {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{}, 0)
	}
	tags[fmt.Sprint(key)] = value

	return context.WithValue(ctx, core.LogTagsKey, tags)
}

func AddNamespace(ctx context.Context, namespaceName string) context.Context {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{}, 0)
	}
	tags["namespace"] = namespaceName

	return context.WithValue(ctx, core.LogTagsKey, tags)
}

func AddInstanceAttr(ctx context.Context, instanceID string, invoker string, callpath string, workflowPath string) context.Context {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{}, 0)
	}
	if trackValue, ok := ctx.Value(core.LogTrackKey).(string); ok {
		tags[string(core.LogTrackKey)] = trackValue
	}

	tags["instance"] = instanceID
	tags["invoker"] = invoker
	tags["callpath"] = callpath
	tags["workflow"] = workflowPath

	return context.WithValue(ctx, core.LogTagsKey, tags)
}

func AddStateAttr(ctx context.Context, state string) context.Context {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{}, 0)
	}
	if trackValue, ok := ctx.Value(core.LogTrackKey).(string); ok {
		tags[string(core.LogTrackKey)] = trackValue
	}

	tags["state"] = state

	return context.WithValue(ctx, core.LogTagsKey, tags)
}

func getCoreAttributes(ctx context.Context) map[string]interface{} {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{})
	}

	if trackValue, ok := ctx.Value(core.LogTrackKey).(string); ok {
		tags[string(core.LogTrackKey)] = trackValue
	}

	return tags
}

func getAttributes(ctx context.Context) map[string]interface{} {
	tags := getCoreAttributes(ctx)
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().TraceID().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()
		tags["trace"] = traceID
		tags["span"] = spanID
	}

	return tags
}

func GetRawLogEntryWithStatus(ctx context.Context, level LogLevel, msg string, status core.LogStatus) map[string]interface{} {
	tags := getAttributes(ctx)
	tags["status"] = status
	tags["level"] = level
	tags["msg"] = msg

	return tags
}

func WithTrack(ctx context.Context, track string) context.Context {
	return context.WithValue(ctx, core.LogTrackKey, track)
}

func BuildNamespaceTrack(namespace string) string {
	return fmt.Sprintf("%v.%v", "namespace", namespace)
}

func BuildInstanceTrack(instance *engine.Instance) string {
	callpath := instance.Instance.ID.String()
	if instance.DescentInfo == nil {
		return fmt.Sprintf("%v.%v", "instance", callpath)
	}
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	return fmt.Sprintf("%v.%v", "instance", callpath)
}

func BuildInstanceTrackViaCallpath(callpath string) string {
	return fmt.Sprintf("%v.%v", "instance", callpath)
}
