package core

import (
	"net/http"
)

const MagicalGatewayNamespace = "gateway_namespace"

type GatewayManager interface {
	http.Handler

	DeleteNamespace(string)
	UpdateNamespace(string)
	UpdateAll()

	GetConsumers(string) ([]*ConsumerBase, error)
	GetRoutes(namespace string) ([]EndpointListItem, error)
}

type EndpointBase struct {
	Methods        []string `json:"methods"         yaml:"methods"`
	PathExtension  string   `json:"path_extension"  yaml:"path_extension"`
	AllowAnonymous bool     `json:"allow_anonymous" yaml:"allow_anonymous"`
	Plugins        Plugins  `json:"plugins"         yaml:"plugins"`
	Timeout        int      `json:"timeout"         yaml:"timeout"`
}

type EndpointListItem struct {
	EndpointBase
	Path     string   `json:"path"`
	Pattern  string   `json:"pattern"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

type ConsumerBase struct {
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	APIKey   string   `json:"api_key"  yaml:"api_key"`
	Tags     []string `json:"tags"     yaml:"tags"`
	Groups   []string `json:"groups"   yaml:"groups"`
}

type Plugins struct {
	Auth     []PluginConfig `json:"auth,omitempty"     yaml:"auth"`
	Inbound  []PluginConfig `json:"inbound,omitempty"  yaml:"inbound"`
	Target   PluginConfig   `json:"target,omitempty"   yaml:"target"`
	Outbound []PluginConfig `json:"outbound,omitempty" yaml:"outbound"`
}

type PluginConfig struct {
	Type          string      `json:"type"          yaml:"type"`
	Configuration interface{} `json:"configuration" yaml:"configuration,omitempty"`
}
