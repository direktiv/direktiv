package functions

var ignoredServiceAnnotations = []string{
	"serving.knative.dev/creator",
	"serving.knative.dev/lastModifier",
	"kubectl.kubernetes.io/last-applied-configuration",
}

var ignoredRevisionAnnotations = []string{
	"serving.knative.dev/creator",
	"serving.knative.dev/lastModifier",
	"serving.knative.dev/routingStateModified",
	"client.knative.dev/updateTimestamp",
}

var ignoredServiceLabels = []string{
	"serving.knative.dev/configurationUID",
	"serving.knative.dev/serviceUID",
}

var ignoredRevisionLabels = []string{
	"serving.knative.dev/configurationUID",
	"serving.knative.dev/serviceUID",
}
