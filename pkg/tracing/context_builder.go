package tracing

import (
	"context"
	"fmt"
	"strings"

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

func getSlogAttributes(ctx context.Context) []interface{} {
	var tags map[string]interface{}

	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{})
	}

	if trackValue, ok := ctx.Value(core.LogTrackKey).(string); ok {
		tags["track"] = trackValue // Add track as a tag to the map
	}
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	tags["trace"] = traceID
	tags["span"] = spanID

	// Convert map back to a slice of key-value pairs
	var result []interface{}
	for k, v := range tags {
		result = append(result, k, v)
	}

	return result
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

func getWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
