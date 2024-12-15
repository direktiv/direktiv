package core

import (
	"fmt"
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

func ParseOpenAPIPathFile(ns string, filePath string, data []byte) Endpoint {
	res := &PathItem{}
	if err := yaml.Unmarshal(data, res); err != nil {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{err.Error()},
		}
	}

	serverPath := ExtractAPIPath(res)
	apiVersion, _ := res.Extensions["direktiv"].(string)
	if !strings.HasPrefix(apiVersion, "api_path") {
		return Endpoint{
			Namespace: ns,
			FilePath:  serverPath,
			Errors:    []string{"invalid api path version"},
		}
	}

	plugins, pluginErr := parsePlugins(res.Extensions["plugins"])
	if pluginErr != nil {
		return Endpoint{
			Namespace: ns,
			FilePath:  serverPath,
			Errors:    []string{pluginErr.Error()},
		}
	}

	allowAnonymous, _ := res.Extensions["allow-anonymous"].(bool)
	if !allowAnonymous && len(plugins.Auth) == 0 {
		return Endpoint{
			Namespace: ns,
			FilePath:  serverPath,
			Errors:    []string{"authentication plugin required but 'allow-anonymous' is false"},
		}
	}

	timeout, _ := res.Extensions["timeout"].(int)
	methods := extractMethods(res)

	return Endpoint{
		Namespace: ns,
		FilePath:  filePath,
		EndpointFile: EndpointFile{
			DirektivAPI:    apiVersion,
			Path:           serverPath,
			PluginsConfig:  plugins,
			AllowAnonymous: allowAnonymous,
			Methods:        methods,
			Timeout:        timeout,
		},
		Errors: []string{},
	}
}

func ExtractAPIPath(res *PathItem) string {
	serverPath := getCleanPath(res.Extensions["path"])

	return serverPath
}

func getCleanPath(rawPath any) string {
	if pathStr, ok := rawPath.(string); ok && pathStr != "" {
		return path.Clean("/" + pathStr)
	}

	return ""
}

func parsePlugins(rawPlugins any) (PluginsConfig, error) {
	plugins := PluginsConfig{}
	pluginsRaw, ok := rawPlugins.(map[string]any)
	if !ok {
		return plugins, fmt.Errorf("missing plugin entry")
	}

	parsePluginField := func(field string, target *[]PluginConfig) {
		if rawField, exists := pluginsRaw[field]; exists {
			if list, ok := rawField.([]any); ok {
				for _, item := range list {
					if pluginConfig, ok := item.(map[string]any); ok {
						config, _ := pluginConfig["configuration"].(map[string]any)
						t, _ := pluginConfig["type"].(string)
						*target = append(*target, PluginConfig{
							Config: config,
							Typ:    t,
						})
					}
				}
			}
		}
	}

	parsePluginField("auth", &plugins.Auth)
	parsePluginField("inbound", &plugins.Inbound)
	parsePluginField("outbound", &plugins.Outbound)

	if targetRaw, exists := pluginsRaw["target"]; exists {
		if targetMap, ok := targetRaw.(map[string]any); ok {
			config, _ := targetMap["configuration"].(map[string]any)
			t, _ := targetMap["type"].(string)
			plugins.Target = PluginConfig{
				Config: config,
				Typ:    t,
			}
		}
	}

	return plugins, nil
}

func extractMethods(res *PathItem) []string {
	methods := []string{}
	if res.Delete != nil {
		methods = append(methods, "DELETE")
	}
	if res.Connect != nil {
		methods = append(methods, "CONNECT")
	}
	if res.Get != nil {
		methods = append(methods, "GET")
	}
	if res.Head != nil {
		methods = append(methods, "HEAD")
	}
	if res.Options != nil {
		methods = append(methods, "OPTIONS")
	}
	if res.Patch != nil {
		methods = append(methods, "PATCH")
	}
	if res.Post != nil {
		methods = append(methods, "POST")
	}
	if res.Put != nil {
		methods = append(methods, "PUT")
	}
	if res.Trace != nil {
		methods = append(methods, "TRACE")
	}

	return methods
}
