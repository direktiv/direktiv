package core

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type GatewayManagerV2 interface {
	http.Handler

	SetEndpoints(list []EndpointV2, cList []ConsumerV2)

	ListEndpoints(namespace string) []EndpointV2
	ListConsumers(namespace string) []ConsumerV2
}

type EndpointFileV2 struct {
	DirektivAPI    string          `yaml:"direktiv_api" json:"-"`
	Methods        []string        `yaml:"methods" json:"methods"`
	Path           string          `yaml:"path" json:"path"` // TODO: fix ui to remove server_path, warnings
	AllowAnonymous bool            `yaml:"allow_anonymous" json:"allow_anonymous"`
	PluginsConfig  PluginsConfigV2 `yaml:"plugins" json:"plugins"`
	Timeout        int             `yaml:"timeout" json:"timeout"`
}

type ConsumerFileV2 struct {
	DirektivAPI string   `yaml:"direktiv_api" json:"-"`
	Username    string   `yaml:"username" json:"username"`
	Password    string   `yaml:"password" json:"password"`
	APIKey      string   `yaml:"api_key" json:"api_key"`
	Tags        []string `yaml:"tags" json:"tags"`
	Groups      []string `yaml:"groups" json:"groups"`
}

type PluginsConfigV2 struct {
	Auth     []PluginConfigV2 `yaml:"auth" json:"auth,omitempty"`
	Inbound  []PluginConfigV2 `yaml:"inbound" json:"inbound,omitempty"`
	Target   PluginConfigV2   `yaml:"target" json:"target,omitempty"`
	Outbound []PluginConfigV2 `yaml:"outbound" json:"outbound,omitempty"`
}

type PluginConfigV2 struct {
	Typ    string         `yaml:"type" json:"type"`
	Config map[string]any `yaml:"configuration" json:"configuration,omitempty"`
}

type PluginV2 interface {
	// NewInstance method creates new plugin instance
	NewInstance(config PluginConfigV2) (PluginV2, error)

	Execute(w http.ResponseWriter, r *http.Request) *http.Request
	Type() string
}

type EndpointV2 struct {
	EndpointFileV2

	Namespace string   `yaml:"-" json:"-"`
	FilePath  string   `yaml:"-" json:"file_path"`
	Errors    []string `yaml:"-" json:"errors"`
}

type ConsumerV2 struct {
	ConsumerFileV2

	Namespace string   `yaml:"-" json:"-"`
	FilePath  string   `yaml:"-" json:"file_path"`
	Errors    []string `yaml:"-" json:"errors"`
}

func ParseConsumerFileV2(data []byte) (*ConsumerFileV2, error) {
	res := &ConsumerFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v1") {
		return nil, fmt.Errorf("invalid consumer api version")
	}

	return res, nil
}

func ParseEndpointFileV2(data []byte) (*EndpointFileV2, error) {
	res := &EndpointFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v1") {
		return nil, fmt.Errorf("invalid endpoint api version")
	}

	return res, nil
}
