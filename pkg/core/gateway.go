package core

import (
	"net/http"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type GatewayManager interface {
	http.Handler

	SetEndpoints(list []Endpoint, cList []Consumer) error
}

type EndpointFile struct {
	DirektivAPI    string        `yaml:"direktiv_api"`
	Methods        []string      `yaml:"methods"`
	Path           string        `yaml:"path"`
	AllowAnonymous bool          `yaml:"allow_anonymous"`
	PluginsConfig  PluginsConfig `yaml:"plugins"`
	Timeout        int           `yaml:"timeout"`
}

type ConsumerFile struct {
	DirektivAPI string   `yaml:"direktiv_api"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	APIKey      string   `yaml:"api_key"`
	Tags        []string `yaml:"tags"`
	Groups      []string `yaml:"groups"`
}

type PluginsConfig struct {
	Auth     []PluginConfig `yaml:"auth"`
	Inbound  []PluginConfig `yaml:"inbound"`
	Target   PluginConfig   `yaml:"target"`
	Outbound []PluginConfig `yaml:"outbound"`
}

type PluginConfig struct {
	Typ    string         `json:"type"                    yaml:"type"`
	Config map[string]any `json:"configuration,omitempty" yaml:"configuration"`
}

type Plugin interface {
	// NewInstance method creates new plugin instance
	NewInstance(config PluginConfig) (Plugin, error)

	Execute(w http.ResponseWriter, r *http.Request) *http.Request
	Type() string
}

type Endpoint struct {
	EndpointFile

	Namespace string
	FilePath  string

	Errors []string
}

type Consumer struct {
	ConsumerFile

	Namespace string
	FilePath  string

	Errors []string
}

func ParseConsumerFile(ns string, filePath string, data []byte) Consumer {
	res := &ConsumerFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return Consumer{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{err.Error()},
		}
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v1") {
		return Consumer{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"invalid consumer api version"},
		}
	}

	return Consumer{
		Namespace:    ns,
		FilePath:     filePath,
		ConsumerFile: *res,
	}
}

func ParseEndpointFile(ns string, filePath string, data []byte) Endpoint {
	res := &EndpointFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{err.Error()},
		}
	}
	if res.Path != "" {
		res.Path = path.Clean("/" + res.Path)
	}
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v1") {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"invalid endpoint api version"},
		}
	}
	if res.PluginsConfig.Target.Typ == "" {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"no target plugin found"},
		}
	}
	if !res.AllowAnonymous && len(res.PluginsConfig.Auth) == 0 {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"no auth plugin configured but 'allow_anonymous' set true"},
		}
	}

	return Endpoint{
		Namespace:    ns,
		FilePath:     filePath,
		EndpointFile: *res,
	}
}
