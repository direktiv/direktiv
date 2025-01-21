package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
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
	Base libopenapi.Document

	Namespace string
	FilePath  string

	Errors []string
}

type Endpoint struct {
	RenderedPathItem openapi3.PathItem

	Config EndpointConfig

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
	fmt.Println("PARSE GATEWAY!!!!")
	gw := Gateway{
		Namespace: ns,
		FilePath:  filePath,
		Errors:    make([]string, 0),
	}

	config := datamodel.DocumentConfiguration{
		AvoidIndexBuild:       true,
		AllowFileReferences:   true,
		AllowRemoteReferences: false,
	}

	// remove paths and servers

	document, err := libopenapi.NewDocumentWithConfiguration(data, &config)
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
	}

	gw.Base = document

	return gw
}

func ParseEndpointFile(ns string, filePath string, data []byte) Endpoint {
	ep := Endpoint{
		Namespace: ns,
		FilePath:  filePath,
		Errors:    make([]string, 0),
	}

	var pathItemMap map[string]interface{}
	err := yaml.Unmarshal(data, &pathItemMap)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	// convert to JSON for openapi library
	b, err := json.Marshal(pathItemMap)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	var pathItem openapi3.PathItem
	err = pathItem.UnmarshalJSON(b)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	api, ok := pathItem.Extensions["x-direktiv-api"]
	if !ok || api != "endpoint/v2" {
		ep.Errors = append(ep.Errors, "invalid endpoint api version")
		return ep
	}

	config, err := parseConfig(&pathItem)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	// add methods
	config.Methods = extractMethods(&pathItem)

	// check for other errors
	if config.Path != "" {
		config.Path = path.Clean("/" + config.Path)
	}

	if config.PluginsConfig.Target.Typ == "" {
		ep.Errors = append(ep.Errors, "no target plugin found")
		return ep
	}

	if !config.AllowAnonymous && len(config.PluginsConfig.Auth) == 0 {
		ep.Errors = append(ep.Errors, "no auth plugin configured but 'allow_anonymous' set false")
		return ep
	}

	ep.Config = config
	ep.RenderedPathItem = pathItem

	return ep
}

func extractMethods(pathItem *openapi3.PathItem) []string {
	methods := make([]string, 0)

	// add methods
	if pathItem.Get != nil {
		methods = append(methods, http.MethodGet)
	}
	if pathItem.Put != nil {
		methods = append(methods, http.MethodPut)
	}
	if pathItem.Post != nil {
		methods = append(methods, http.MethodPost)
	}
	if pathItem.Delete != nil {
		methods = append(methods, http.MethodDelete)
	}
	if pathItem.Options != nil {
		methods = append(methods, http.MethodOptions)
	}
	if pathItem.Head != nil {
		methods = append(methods, http.MethodHead)
	}
	if pathItem.Patch != nil {
		methods = append(methods, http.MethodPatch)
	}
	if pathItem.Trace != nil {
		methods = append(methods, http.MethodTrace)
	}

	return methods
}

func parseConfig(pathItem *openapi3.PathItem) (EndpointConfig, error) {
	var config EndpointConfig

	c, ok := pathItem.Extensions["x-direktiv-config"]
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
