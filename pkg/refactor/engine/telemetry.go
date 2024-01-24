package engine

import (
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
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
	tags["revision-id"] = instance.Instance.RevisionID.String()

	tags["namespace"] = instance.TelemetryInfo.NamespaceName

	return tags
}

func (instance *Instance) GetSlogAttributes() []interface{} {
	tags := make([]interface{}, 0)

	rootInstanceID := instance.Instance.ID
	callpath := ""
	if len(instance.DescentInfo.Descent) > 0 {
		rootInstanceID = instance.DescentInfo.Descent[0].ID
	}
	for _, v := range instance.DescentInfo.Descent {
		callpath += "/" + v.ID.String()
	}

	tags = append(tags, "stream", fmt.Sprintf("%v.%v", recipient.Instance, rootInstanceID))
	tags = append(tags, "instance-id", instance.Instance.ID)
	tags = append(tags, "invoker", instance.Instance.Invoker)
	tags = append(tags, "callpath", callpath)
	tags = append(tags, "workflow", instance.Instance.WorkflowPath)
	tags = append(tags, "namespace", instance.Instance.Namespace)
	tags = append(tags, "source", recipient.Instance)
	tags = append(tags, "root-instance-id", rootInstanceID)
	tags = append(tags, "trace", instance.TelemetryInfo.TraceID)
	tags = append(tags, "span", instance.TelemetryInfo.SpanID)
	// tags = append(tags, "callpath", instance.TelemetryInfo.CallPath) // Todo this is value is not filled

	return tags
}

func getWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
