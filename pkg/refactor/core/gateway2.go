package core

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	GatewayCtxKeyConsumers      = "ctx_consumers"
	GatewayCtxKeyActiveConsumer = "ctx_active_consumer"
)

type GatewayManagerV2 interface {
	http.Handler

	SetEndpoints(list []EndpointV2, cList []ConsumerV2)

	ListEndpoints(namespace string) []EndpointV2
	ListConsumers(namespace string) []ConsumerV2
}

type EndpointFileV2 struct {
	DirektivAPI    string          `yaml:"direktiv_api"`
	Methods        []string        `yaml:"methods"`
	Path           string          `yaml:"path"`
	AllowAnonymous bool            `yaml:"allow_anonymous"`
	PluginsConfig  PluginsConfigV2 `yaml:"plugins"`
	Timeout        int             `yaml:"timeout"`
}

type ConsumerFileV2 struct {
	DirektivAPI string   `yaml:"direktiv_api"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	APIKey      string   `yaml:"api_key"`
	Tags        []string `yaml:"tags"`
	Groups      []string `yaml:"groups"`
}

type PluginsConfigV2 struct {
	Auth     []PluginConfigV2 `yaml:"auth"`
	Inbound  []PluginConfigV2 `yaml:"inbound"`
	Target   PluginConfigV2   `yaml:"target"`
	Outbound []PluginConfigV2 `yaml:"outbound"`
}

type PluginConfigV2 struct {
	Typ    string         `yaml:"type"`
	Config map[string]any `yaml:"configuration"`
}

type PluginV2 interface {
	// NewInstance method creates new plugin instance
	NewInstance(config PluginConfigV2) (PluginV2, error)

	Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error)
	Type() string
}

type EndpointV2 struct {
	EndpointFileV2

	Namespace string
	FilePath  string
	Errors    []error
}

type ConsumerV2 struct {
	ConsumerFileV2

	Namespace string
	FilePath  string
	Errors    []error
}

func ParseConsumerFileV2(data []byte) (*ConsumerFileV2, error) {
	res := &ConsumerFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v2") {
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
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v2") {
		return nil, fmt.Errorf("invalid endpoint api version")
	}

	return res, nil
}
