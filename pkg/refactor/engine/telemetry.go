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

	rootInstanceID := instance.Instance.ID
	callpath := ""
	if len(instance.DescentInfo.Descent) > 0 {
		rootInstanceID = instance.DescentInfo.Descent[0].ID
	}
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	tags = append(tags, "instance-id", instance.Instance.ID)
	tags = append(tags, "invoker", instance.Instance.Invoker)
	tags = append(tags, "callpath", callpath)
	tags = append(tags, "workflow", instance.Instance.WorkflowPath)
	tags = append(tags, "namespace", instance.Instance.Namespace)
	tags = append(tags, "root-instance-id", rootInstanceID)

	if trackValue, ok := ctx.Value(core.TrackKey).(string); ok {
		tags = append(tags, "track", trackValue)
	}

	return context.WithValue(ctx, core.TagsKey, tags)
}

func GetSlogAttributes(ctx context.Context) []interface{} {
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
	tags := GetSlogAttributes(ctx)
	tags = append(tags, "status", status)

	return tags
}

func GetSlogAttributesWithError(ctx context.Context, err error) []interface{} {
	tags := GetSlogAttributes(ctx)
	tags = append(tags, "error", err)
	tags = append(tags, "status", "error")

	return tags
}

func WithTrack(ctx context.Context, track string) context.Context {
	return context.WithValue(ctx, core.TrackKey, track)
}

func buildNamespaceTrack(namespace string) string {
	return fmt.Sprintf("%v.%v", "namespace", namespace)
}

func buildInstanceTrack(instance *Instance) string {
	callpath := ""
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	return fmt.Sprintf("%v.%v", "instance", callpath)
}

func getWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
