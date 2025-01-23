package core

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	v3low "github.com/pb33f/libopenapi/datamodel/low/v3"
	"github.com/pb33f/libopenapi/index"
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
	PathItem v3high.PathItem
	Doc      libopenapi.Document

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

type DummyFS struct {
}

type DummyFSFile struct {
}

func (d DummyFSFile) Close() error {
	return nil
}

func (d DummyFSFile) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (d DummyFSFile) Stat() (fs.FileInfo, error) {
	return fs.FileInfo{}, io.EOF
}

func (d *DummyFS) Open(name string) (fs.File, error) {
	fmt.Println("FFFFFFFFFFFFFFFFFFFFFFEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE")
	fmt.Println(name)
	return DummyFSFile{}, nil
}

func (d *DummyFS) GetFiles() map[string]index.RolodexFile {
	return make(map[string]index.RolodexFile)
}

func ParseEndpointFile(ns string, filePath string, data []byte) Endpoint {
	ep := Endpoint{
		Namespace: ns,
		FilePath:  filePath,
		Errors:    make([]string, 0),
	}

	docConfig := datamodel.DocumentConfiguration{
		AvoidIndexBuild:       true,
		AllowFileReferences:   true,
		AllowRemoteReferences: false,
	}

	// can not throw an error, fake document for root index
	doc, _ := libopenapi.NewDocumentWithConfiguration([]byte("openapi: 3.0.0\ninfo:\n   version: \"1.0\"\n   title: dummy\npaths: {}\n"), &docConfig)
	highDoc, _ := doc.BuildV3Model()

	doc.GetRolodex().AddLocalFS("/", &DummyFS{})

	var interim map[string]interface{}
	err := yaml.Unmarshal(data, &interim)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	var node yaml.Node
	err = node.Encode(interim)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	var lowPathItem v3low.PathItem
	err = lowPathItem.Build(context.Background(), nil, &node, doc.GetRolodex().GetRootIndex())
	if err != nil {
		fmt.Println(err)
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	pathItem := v3high.NewPathItem(&lowPathItem)

	api, found := pathItem.Extensions.Get("x-direktiv-api")
	if !found || api.Value != "endpoint/v2" {
		ep.Errors = append(ep.Errors, "invalid endpoint api version")
		return ep
	}

	config, err := parseConfig(pathItem)
	if err != nil {
		ep.Errors = append(ep.Errors, err.Error())
		return ep
	}

	// add methods
	config.Methods = extractMethods(pathItem)

	// check for other errors
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

	fmt.Println("PARSE ENDPOINT5")
	highDoc.Model.Paths.PathItems.Set(config.Path, pathItem)
	_, newDoc, _, errs := doc.RenderAndReload()
	if len(errs) > 0 {
		for i := range errs {
			ep.Errors = append(ep.Errors, errs[i].Error())
		}
		return ep
	}

	fmt.Println("PARSED ENDPOINT")
	fmt.Println(ep.Errors)

	ep.Config = config
	ep.PathItem = *pathItem
	ep.Doc = newDoc

	return ep
}

func extractMethods(pathItem *v3high.PathItem) []string {
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

func parseConfig(pathItem *v3high.PathItem) (EndpointConfig, error) {
	var config EndpointConfig

	c, found := pathItem.Extensions.Get("x-direktiv-config")

	// c, ok := m["x-direktiv-config"]
	if !found {
		return config, fmt.Errorf("no endpoint configuration found")
	}

	ct, err := yaml.Marshal(c)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(ct, &config)

	return config, err
}
