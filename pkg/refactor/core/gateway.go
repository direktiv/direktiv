package core

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
)

const MagicalGatewayNamespace = "gateway_namespace"

type GatewayManager interface {
	http.Handler

	DeleteNamespace(string)
	UpdateNamespace(string)
	UpdateAll()
}

type Endpoint struct {
	EndpointFile            *spec.EndpointFile
	Namespace               string
	FilePath                string
	AuthPluginInstances     []plugins.PluginInstance
	InboundPluginInstances  []plugins.PluginInstance
	TargetPluginInstance    plugins.PluginInstance
	OutboundPluginInstances []plugins.PluginInstance
	Errors                  []string
	Warnings                []string
}
