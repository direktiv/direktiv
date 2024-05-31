package engine

import (
	"context"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/flow/nohome/recipient"
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
	tags, ok := ctx.Value(core.LogTagsKey).(map[string]interface{})
	if !ok {
		tags = make(map[string]interface{}, 0)
	}

	callpath := ""

	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	tags["instance"] = instance.Instance.ID
	tags["invoker"] = instance.Instance.Invoker
	tags["callpath"] = callpath
	tags["workflow"] = instance.Instance.WorkflowPath

	return context.WithValue(ctx, core.LogTagsKey, tags)
}
func getWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
