package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type GatewayManager interface {
	http.Handler

	SetEndpoints(list []Endpoint, cList []Consumer, gList []Gateway) error
}

type EndpointConfig struct {
	Methods        []string      `yaml:"methods"`
	Path           string        `yaml:"path"`
	AllowAnonymous bool          `yaml:"allow_anonymous"`
	PluginsConfig  PluginsConfig `yaml:"plugins"`
	Timeout        int           `yaml:"timeout"`
	SkipOpenAPI    bool          `yaml:"skip_openapi"`
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

	Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request)
	Type() string
}

type Gateway struct {
	Base []byte

	Namespace string
	FilePath  string
	IsVirtual bool

	Errors []string
}

type Endpoint struct {
	Config EndpointConfig
	Base   []byte

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

	// remove paths and servers
	var interim map[string]interface{}
	err := yaml.Unmarshal(data, &interim)
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
	}

	// these will be generated in direktiv based on route files
	delete(interim, "paths")
	delete(interim, "servers")

	out, err := yaml.Marshal(interim)
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
	}

	// set cleaned up spec
	gw.Base = out

	return gw
}

func ParseEndpointFile(ns string, filePath string, data []byte) Endpoint {
	ep := Endpoint{
		Namespace: ns,
		FilePath:  filePath,
		Errors:    make([]string, 0),
	}

	var interim map[string]interface{}
	err := yaml.Unmarshal(data, &interim)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	jsonData, err := json.Marshal(interim)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}
	ep.Base = jsonData

	config, err := parseConfig(interim)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	// add methods
	config.Methods = extractMethods(interim)

	// check for other errors
	if config.Path != "" {
		config.Path = path.Clean("/" + config.Path)
	} else {
		ep.Errors = append(ep.Errors, "no path for route specified")
	}

	if len(config.Methods) == 0 {
		ep.Errors = append(ep.Errors, "no valid http method available")
	}

	if config.PluginsConfig.Target.Typ == "" {
		ep.Errors = append(ep.Errors, "no target plugin found")
	}

	if !config.AllowAnonymous && len(config.PluginsConfig.Auth) == 0 {
		ep.Errors = append(ep.Errors, "no auth plugin configured but 'allow_anonymous' set false")
	}

	ep.Config = config

	return ep
}

func extractMethods(pathItem map[string]interface{}) []string {
	methods := make([]string, 0)

	availableMethods := []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodPost,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodHead,
		http.MethodPatch,
		http.MethodTrace,
	}

	for i := range availableMethods {
		if _, ok := pathItem[strings.ToLower(availableMethods[i])]; ok {
			methods = append(methods, availableMethods[i])
		}
	}

	return methods
}

func parseConfig(pathItem map[string]interface{}) (EndpointConfig, error) {
	var config EndpointConfig

	c, ok := pathItem["x-direktiv-config"]
	if !ok {
		return config, fmt.Errorf("no endpoint configuration found")
	}

	ct, err := yaml.Marshal(c)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(ct, &config)

	return config, err
}
