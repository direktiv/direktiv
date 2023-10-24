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

func GetKnativeServiceURL(knativeNamespace string, namespace string, typ string, file string, name string) string {
	id := (&core.ServiceConfig{
		Typ:       typ,
		Namespace: namespace,
		FilePath:  file,
		Name:      name,
	}).GetID()

	return fmt.Sprintf("http://%s.%s.svc.cluster.local", id, knativeNamespace)
}

func GetDockerServiceURL(namespace string, typ string, file string, name string) string {
	id := (&core.ServiceConfig{
		Typ:       typ,
		Namespace: namespace,
		FilePath:  file,
		Name:      name,
	}).GetID()

	return fmt.Sprintf("http://%s", id)
}

var GetServiceURL func(namespace string, typ string, file string, name string) string
