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
	FilePath                string
	AuthPluginInstances     []plugins.PluginInstance
	InboundPluginInstances  []plugins.PluginInstance
	OutboundPluginInstances []plugins.PluginInstance
}

// type Plugins struct {
// 	Auth     []spec.PluginConfig `yaml:"auth"`
// 	Inboud   []spec.PluginConfig `yaml:"inbound"`
// 	Target   []spec.PluginConfig `yaml:"target"`
// 	Outbound []spec.PluginConfig `yaml:"outbound"`
// }

// type Endpoint struct {
// 	Methods []string `json:"methods"`
// 	// FilePath       string   `json:"file_path"`
// 	// PathExtension  string   `yaml:"path_extension"`
// 	AllowAnonymous bool     `yaml:"allow_anonymous"`
// 	Plugins        Plugins  `yaml:"plugins"`
// 	Errors         []string `json:"errors"`
// }

// type Consumer struct {
// 	Username string   `yaml:"username"`
// 	Password string   `yaml:"password"`
// 	APIKey   string   `yaml:"api_key"`
// 	Tags     []string `yaml:"tags"`
// 	Groups   []string `yaml:"groups"`
// }
