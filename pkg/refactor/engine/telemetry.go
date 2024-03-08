package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"go.opentelemetry.io/otel/trace"
)

func (instance *Instance) GetAttributes(recipientType recipient.RecipientType) map[string]string {
	tags := make(map[string]string)
	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}
	tags["recipientType"] = string(recipientType)
	tags["instance-id"] = instance.Instance.ID.String()
	tags["invoker"] = instance.Instance.Invoker
	tags["callpath"] = callpath
	tags["workflow"] = getWorkflow(instance.Instance.WorkflowPath)
	tags["namespace-id"] = instance.Instance.NamespaceID.String()

	tags["namespace"] = instance.TelemetryInfo.NamespaceName

	return tags
}

func (instance *Instance) WithTags(ctx context.Context) context.Context {
	tags, ok := ctx.Value(core.TagsKey).([]interface{})
	if !ok {
		tags = make([]interface{}, 0)
	}

	callpath := ""

	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	tags = append(tags, "instance", instance.Instance.ID)
	tags = append(tags, "invoker", instance.Instance.Invoker) // TODO: value is empty.
	tags = append(tags, "callpath", callpath)
	tags = append(tags, "workflow", instance.Instance.WorkflowPath) // TODO: value is empty.

	return context.WithValue(ctx, core.TagsKey, tags)
}

func AddTag(ctx context.Context, key, value interface{}) context.Context {
	tags, ok := ctx.Value(core.TagsKey).([]interface{})
	if !ok {
		tags = make([]interface{}, 0)
	}
	tags = append(tags, key, value)

	return context.WithValue(ctx, core.TagsKey, tags)
}

func getSlogAttributes(ctx context.Context) []interface{} {
	tags, ok := ctx.Value(core.TagsKey).([]interface{})
	if !ok {
		tags = make([]interface{}, 0)
	}
	if trackValue, ok := ctx.Value(core.TrackKey).(string); ok {
		tags = append(tags, "track", trackValue)
	}

	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	tags = append(tags, "trace", traceID)
	tags = append(tags, "span", spanID)

	return tags
}

func GetSlogAttributesWithStatus(ctx context.Context, status core.Status) []interface{} {
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
	return context.WithValue(ctx, core.TrackKey, track)
}

func BuildNamespaceTrack(namespace string) string {
	return fmt.Sprintf("%v.%v", "namespace", namespace)
}

func BuildInstanceTrack(instance *Instance) string {
	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}
	if callpath == "" {
		callpath = instance.Instance.ID.String()
	}

	return fmt.Sprintf("%v.%v", "instance", callpath)
}

func getWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
