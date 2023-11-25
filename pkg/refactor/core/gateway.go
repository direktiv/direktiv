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

	GetConsumers(string) ([]*spec.ConsumerFile, error)
	GetRoutes(namespace string) ([]EndpointListItem, error)
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

type EndpointListItem struct {
	spec.EndpointFile
	Path     string   `json:"path"`
	Pattern  string   `json:"pattern"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}
