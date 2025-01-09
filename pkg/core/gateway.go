package core

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

type GatewayManager interface {
	http.Handler

	SetEndpoints(list []Endpoint, cList []Consumer, baseDefs []Gateway) error
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

type Gateway struct {
	RenderedBase openapi3.T

	Namespace string
	FilePath  string

	Errors []string
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

func ParseGatewayFile(ns string, filePath string, data []byte) Gateway {
	gw := Gateway{
		Namespace: ns,
		FilePath:  filePath,
		Errors:    make([]string, 0),
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		return nil, nil
	}

	base, err := loader.LoadFromData(data)
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
		return gw
	}

	// remove paths and server because it will be generated
	base.Paths = openapi3.NewPaths()
	base.Servers = openapi3.Servers{}
	gw.RenderedBase = *base

	err = base.Validate(context.Background())
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
		return gw
	}

	return gw
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
			Errors:    []string{"no auth plugin configured but 'allow_anonymous' set false"},
		}
	}

	return Endpoint{
		Namespace:    ns,
		FilePath:     filePath,
		EndpointFile: *res,
	}
}
