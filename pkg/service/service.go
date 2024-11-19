package service

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
)

const (
	// httpsProxy = "HTTPS_PROXY"
	// httpProxy  = "HTTP_PROXY"
	// noProxy    = "NO_PROXY".

	containerUser        = "direktiv-container"
	containerSidecar     = "direktiv-sidecar"
	containerSidecarPort = 80
)

// GetServiceURL is a global function that know how to construct a service url based on service parameters.
// You need to call SetupGetServiceURLFunc function to construct GetServiceURL.
var GetServiceURL func(namespace string, typ string, file string, name string) string

func getKnativeServiceURL(config *core.Config, namespace string, typ string, file string, name string) string {
	// Construct the service ID
	serviceID := (&core.ServiceFileData{
		Typ:       typ,
		Namespace: namespace,
		FilePath:  file,
		Name:      name,
	}).GetID()

	if config.IngressHost != "" {
		// Construct external URL for ingress host
		return fmt.Sprintf("http://%s.%s.%s.nip.io", serviceID, config.KnativeNamespace, config.IngressHost)
	}

	// Default to internal cluster-local DNS
	return fmt.Sprintf("http://%s.%s.svc.cluster.local", serviceID, config.KnativeNamespace)
}

func SetupGetServiceURLFunc(config *core.Config) {
	GetServiceURL = func(namespace string, typ string, file string, name string) string {
		return getKnativeServiceURL(config, namespace, typ, file, name)
	}
}
