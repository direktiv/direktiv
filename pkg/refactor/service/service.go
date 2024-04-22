package service

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

const (
	// httpsProxy = "HTTPS_PROXY"
	// httpProxy  = "HTTP_PROXY"
	// noProxy    = "NO_PROXY".

	containerUser        = "direktiv-container"
	containerSidecar     = "direktiv-sidecar"
	containerSidecarPort = 8890
)

// GetServiceURL is a global function that know how to construct a service url based on service parameters.
// You need to call SetupGetServiceURLFunc function to construct GetServiceURL.
var GetServiceURL func(namespace string, typ string, file string, name string) string

func getKnativeServiceURL(knativeNamespace string, namespace string, typ string, file string, name string) string {
	id := (&core.ServiceFileData{
		Typ:       typ,
		Namespace: namespace,
		FilePath:  file,
		Name:      name,
	}).GetID()

	return fmt.Sprintf("http://%s.%s.svc.cluster.local", id, knativeNamespace)
}

func getDockerServiceURL(namespace string, typ string, file string, name string) string {
	id := (&core.ServiceFileData{
		Typ:       typ,
		Namespace: namespace,
		FilePath:  file,
		Name:      name,
	}).GetID()

	return fmt.Sprintf("http://%s", id)
}

func SetupGetServiceURLFunc(config *core.Config, withDocker bool) {
	GetServiceURL = func(namespace string, typ string, file string, name string) string {
		if withDocker {
			return getDockerServiceURL(namespace, typ, file, name)
		}

		return getKnativeServiceURL(config.KnativeNamespace, namespace, typ, file, name)
	}
}
