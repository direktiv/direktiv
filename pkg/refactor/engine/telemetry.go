package engine

import (
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
	tags["workflow"] = getWorkflow(instance.Instance.CalledAs)
	tags["workflow-id"] = instance.Instance.WorkflowID.String()
	tags["namespace-id"] = instance.Instance.NamespaceID.String()
	tags["revision-id"] = instance.Instance.RevisionID.String()

	tags["namespace"] = instance.TelemetryInfo.NamespaceName

	return tags
}

func getWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
