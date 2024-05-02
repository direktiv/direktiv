package core

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	GATEWAY_CTX_KEY_CONSUMERS       = "ctx_consumers"
	GATEWAY_CTX_KEY_ACTIVE_CONSUMER = "ctx_active_consumer"
)

type GatewayManagerV2 interface {
	http.Handler

	SetEndpoints(list []EndpointV2)
	SetConsumers(list []ConsumerV2)

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
	Auth     []PluginConfigV2
	Inbound  []PluginConfigV2
	Target   PluginConfigV2
	Outbound []PluginConfigV2
}

type PluginConfigV2 struct {
	Typ    string
	Config map[string]any
}

type PluginV2 interface {
	Execute(w http.ResponseWriter, r *http.Request) *http.Request
	Config() any
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
}

func FindConsumerByUser(user string, list []ConsumerV2) *ConsumerV2 {
	for _, item := range list {
		if item.Username == user {
			return &item
		}
	}

	return nil
}

func ParseConsumerFileV2(data []byte) (*ConsumerFileV2, error) {
	res := &ConsumerFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v2") {
		return nil, fmt.Errorf("invalid axiliary api version")
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
		return nil, fmt.Errorf("invalid route api version")
	}

	return res, nil
}
