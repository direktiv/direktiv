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

// TODO test convert here.
func ParseOpenAPIPathFile(ns string, filePath string, data []byte) Endpoint {
	res := &PathItem{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{err.Error()},
		}
	}
	if filePath != "" {
		filePath = path.Clean("/" + filePath)
	}
	apiVersion, _ := res.Extensions["direktiv"].(string)
	if !strings.HasPrefix(apiVersion, "api_path") {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"invalid api path version"},
		}
	}
	fmt.Printf("Plugins field: %+v\n", res.Extensions["plugins"])

	// Retrieve the plugins data (you've already parsed it into a map)
	pluginsRaw, ok := res.Extensions["plugins"].(map[string]any)
	if !ok {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"missing plugin entry"},
		}
	}

	// Initialize PluginsConfig (struct to hold the processed plugins)
	plugins := PluginsConfig{}

	// Handle 'auth' field
	if authRaw, exists := pluginsRaw["auth"]; exists {
		authList, ok := authRaw.([]any)
		if ok {
			for _, item := range authList {
				pluginConfig, ok := item.(map[string]any)
				if ok {
					plugins.Auth = append(plugins.Auth, PluginConfig{Config: pluginConfig})
				}
			}
		}
	}

	// Handle 'inbound' field
	if inboundRaw, exists := pluginsRaw["inbound"]; exists {
		inboundList, ok := inboundRaw.([]any)
		if ok {
			for _, item := range inboundList {
				pluginConfig, ok := item.(map[string]any)
				if ok {
					plugins.Inbound = append(plugins.Inbound, PluginConfig{Config: pluginConfig})
				}
			}
		}
	}

	// Handle 'outbound' field
	if outboundRaw, exists := pluginsRaw["outbound"]; exists {
		outboundList, ok := outboundRaw.([]any)
		if ok {
			for _, item := range outboundList {
				pluginConfig, ok := item.(map[string]any)
				if ok {
					plugins.Outbound = append(plugins.Outbound, PluginConfig{Config: pluginConfig})
				}
			}
		}
	}

	// Handle 'target' field separately (it’s a single map and needs to be unmarshalled)
	if targetRaw, exists := pluginsRaw["target"]; exists {
		targetMap, ok := targetRaw.(map[string]any)
		if ok {
			plugins.Target = PluginConfig{Config: targetMap} // Just assign the map for 'target'
		}
	}

	allowAnonymous, _ := res.Extensions["allow-anonymous"].(bool)
	if !allowAnonymous && len(plugins.Auth) == 0 {
		return Endpoint{
			Namespace: ns,
			FilePath:  filePath,
			Errors:    []string{"authentication plugin required but 'allow-anonymous' is false"},
		}
	}

	timeout, _ := res.Extensions["timeout"].(int)

	methods := []string{}
	if res.Delete != nil {
		methods = append(methods, "delete")
	}
	if res.Connect != nil {
		methods = append(methods, "connect")
	}
	if res.Get != nil {
		methods = append(methods, "get")
	}
	if res.Head != nil {
		methods = append(methods, "head")
	}
	if res.Options != nil {
		methods = append(methods, "options")
	}
	if res.Patch != nil {
		methods = append(methods, "patch")
	}
	if res.Post != nil {
		methods = append(methods, "post")
	}
	if res.Put != nil {
		methods = append(methods, "put")
	}
	if res.Trace != nil {
		methods = append(methods, "trace")
	}

	return Endpoint{
		Namespace: ns,
		FilePath:  filePath,
		EndpointFile: EndpointFile{
			DirektivAPI:    apiVersion,
			Path:           filePath,
			PluginsConfig:  plugins,
			AllowAnonymous: allowAnonymous,
			Methods:        methods,
			Timeout:        timeout, // TODO: timeout via spec?
		},
	}
}
