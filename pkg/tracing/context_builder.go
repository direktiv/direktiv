package tracing

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
	"go.opentelemetry.io/otel/trace"
)

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
		tags["track"] = trackValue
	}

	tags["instance"] = instanceID
	tags["invoker"] = invoker
	tags["callpath"] = callpath
	tags["workflow"] = workflowPath

	return context.WithValue(ctx, core.LogTagsKey, tags)
}

func AddTraceAttr(ctx context.Context, traceID, spanID string) context.Context {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{}, 0)
	}
	if trackValue, ok := ctx.Value(core.LogTrackKey).(string); ok {
		tags["track"] = trackValue
	}

	tags["trace"] = traceID
	tags["span"] = spanID

	return context.WithValue(ctx, core.LogTagsKey, tags)
}

func getSlogAttributes(ctx context.Context) []interface{} {
	tags := getAttributes(ctx)

	// Convert map back to a slice of key-value pairs
	var result []interface{}
	for k, v := range tags {
		result = append(result, k, v)
	}

	return result
}

func getAttributes(ctx context.Context) map[string]interface{} {
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{})
	}

	if trackValue, ok := ctx.Value(core.LogTrackKey).(string); ok {
		tags["track"] = trackValue
	}
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	tags["trace"] = traceID
	tags["span"] = spanID

	return tags
}

func GetLogEntryWithStatus(ctx context.Context, level string, msg string, status core.LogStatus) map[string]interface{} {
	tags := getAttributes(ctx)
	tags["status"] = status
	tags["level"] = level

	return tags
}

func GetLogEntryWithError(ctx context.Context, msg string, err error) map[string]interface{} {
	tags := getAttributes(ctx)
	tags["error"] = err
	tags["status"] = "error"
	tags["level"] = "error"

	return tags
}

func GetSlogAttributesWithStatus(ctx context.Context, status core.LogStatus) []interface{} {
	tags := getSlogAttributes(ctx)
	tags = append(tags, "status", status)

	return tags
}

func GetSlogAttributesWithError(ctx context.Context, err error) []interface{} {
	tags := getSlogAttributes(ctx)
	tags = append(tags, "error", err)
	tags = append(tags, "status", "error")

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
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	return fmt.Sprintf("%v.%v", "instance", callpath)
}
