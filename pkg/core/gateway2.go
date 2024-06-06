package core

import (
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type GatewayManagerV2 interface {
	http.Handler

	SetEndpoints(list []EndpointV2, cList []ConsumerV2)
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
	Typ    string         `json:"type"                    yaml:"type"`
	Config map[string]any `json:"configuration,omitempty" yaml:"configuration"`
}

type PluginV2 interface {
	// NewInstance method creates new plugin instance
	NewInstance(config PluginConfigV2) (PluginV2, error)

	Execute(w http.ResponseWriter, r *http.Request) *http.Request
	Type() string
}

type EndpointV2 struct {
	EndpointFileV2

	Namespace string
	FilePath  string

	Errors []string
}

type ConsumerV2 struct {
	ConsumerFileV2

	Namespace string
	FilePath  string

	Errors []string
}

func ParseConsumerFileV2(ns string, filePath string, data []byte) ConsumerV2 {
	res := &ConsumerFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return ConsumerV2{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{err.Error()},
		}
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v1") {
		return ConsumerV2{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"invalid consumer api version"},
		}
	}

	return ConsumerV2{
		Namespace:      ns,
		FilePath:       filePath,
		ConsumerFileV2: *res,
	}
}

func ParseEndpointFileV2(ns string, filePath string, data []byte) EndpointV2 {
	res := &EndpointFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return EndpointV2{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{err.Error()},
		}
	}
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v1") {
		return EndpointV2{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"invalid endpoint api version"},
		}
	}
	if res.PluginsConfig.Target.Typ == "" {
		return EndpointV2{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"no target plugin found"},
		}
	}
	if !res.AllowAnonymous && len(res.PluginsConfig.Auth) == 0 {
		return EndpointV2{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"no auth plugin configured but 'allow_anonymous' set true"},
		}
	}

	return EndpointV2{
		Namespace:      ns,
		FilePath:       filePath,
		EndpointFileV2: *res,
	}
}
