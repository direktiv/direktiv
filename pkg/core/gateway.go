package core

import (
	"fmt"
	"net/http"
	"path"
	"strings"

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
	Config EndpointConfig
	Yaml   map[string]interface{}

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

	config := datamodel.DocumentConfiguration{
		AvoidIndexBuild:       true,
		AllowFileReferences:   true,
		AllowRemoteReferences: false,
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

	document, err := libopenapi.NewDocumentWithConfiguration(out, &config)
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

	// docConfig := datamodel.DocumentConfiguration{
	// 	AvoidIndexBuild:       true,
	// 	AllowFileReferences:   true,
	// 	AllowRemoteReferences: false,
	// 	LocalFS:               &DummyFS{
	// 		// fileStore: fileStore,
	// 		// ns:        ns,
	// 	},
	// }

	// can not throw an error, fake document for root index
	// doc, _ := libopenapi.NewDocumentWithConfiguration([]byte("openapi: 3.0.0\ninfo:\n   version: \"1.0\"\n   title: dummy\npaths: {}\n"), &docConfig)
	// highDoc, _ := doc.BuildV3Model()

	// doc.GetRolodex().AddLocalFS("/", &DummyFS{})

	var interim map[string]interface{}
	err := yaml.Unmarshal(data, &interim)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	// var node yaml.Node
	// err = node.Encode(interim)
	// if err != nil {
	// 	ep.Errors = append(ep.Errors, err.Error())
	// 	return ep
	// }

	// var lowPathItem v3low.PathItem
	// err = lowPathItem.Build(context.Background(), nil, &node, doc.GetRolodex().GetRootIndex())
	// if err != nil && !strings.Contains(err.Error(), "cannot be found at line") {
	// 	fmt.Println("111ERROROROR HERERERERERRE")
	// 	fmt.Println(reflect.TypeOf(err))
	// 	fmt.Println(err.Error())
	// 	ep.Errors = append(ep.Errors, err.Error())
	// 	return ep
	// }

	// pathItem := v3high.NewPathItem(&lowPathItem)

	api, found := interim["x-direktiv-api"]
	if !found || api != "endpoint/v2" {
		ep.Errors = append(ep.Errors, "invalid endpoint api version")
		return ep
	}

	config, err := parseConfig(interim)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	// add methods
	config.Methods = extractMethods(interim)

	// // check for other errors
	if config.Path != "" {
		config.Path = path.Clean("/" + config.Path)
	} else {
		ep.Errors = append(ep.Errors, "no path for route specified")
		return ep
	}

	if config.PluginsConfig.Target.Typ == "" {
		ep.Errors = append(ep.Errors, "no target plugin found")
		return ep
	}

	if !config.AllowAnonymous && len(config.PluginsConfig.Auth) == 0 {
		ep.Errors = append(ep.Errors, "no auth plugin configured but 'allow_anonymous' set false")
		return ep
	}

	// fmt.Println("PARSE ENDPOINT5")
	// highDoc.Model.Paths.PathItems.Set(config.Path, pathItem)
	// _, newDoc, _, errs := doc.RenderAndReload()
	// if len(errs) > 0 {
	// 	for i := range errs {
	// 		ep.Errors = append(ep.Errors, errs[i].Error())
	// 	}
	// 	return ep
	// }

	// fmt.Println("PARSED ENDPOINT")
	// fmt.Println(ep.Errors)

	ep.Config = config
	ep.Yaml = interim
	// ep.PathItem = *pathItem
	// ep.Doc = newDoc

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
		if _, ok := pathItem[availableMethods[i]]; ok {
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
