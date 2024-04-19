package core

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type GatewayManager interface {
	http.Handler

	DeleteNamespace(namespace string)
	UpdateNamespace(namespace string)
	UpdateAll()

	GetConsumers(namespace string) ([]*ConsumerFile, error)
	GetRoutes(namespace string, filteredPath string) ([]*Endpoint, error)
}

type EndpointFile struct {
	DirektivAPI    string   `json:"-"               yaml:"direktiv_api"`
	Methods        []string `json:"methods"         yaml:"methods"`
	Path           string   `json:"path"            yaml:"path"`
	AllowAnonymous bool     `json:"allow_anonymous" yaml:"allow_anonymous"`
	Plugins        Plugins  `json:"plugins"         yaml:"plugins"`
	Timeout        int      `json:"timeout"         yaml:"timeout"`
}

type ConsumerFile struct {
	DirektivAPI string   `json:"-"        yaml:"direktiv_api"`
	Username    string   `json:"username" yaml:"username"`
	Password    string   `json:"password" yaml:"password"`
	APIKey      string   `json:"api_key"  yaml:"api_key"`
	Tags        []string `json:"tags"     yaml:"tags"`
	Groups      []string `json:"groups"   yaml:"groups"`
}

type Plugins struct {
	Auth     []PluginConfig `json:"auth,omitempty"     yaml:"auth"`
	Inbound  []PluginConfig `json:"inbound,omitempty"  yaml:"inbound"`
	Target   *PluginConfig  `json:"target,omitempty"   yaml:"target"`
	Outbound []PluginConfig `json:"outbound,omitempty" yaml:"outbound"`
}

type PluginConfig struct {
	Type          string                 `json:"type,omitempty"          yaml:"type"`
	Configuration map[string]interface{} `json:"configuration,omitempty" yaml:"configuration"`
}

type PluginInstance interface {
	ExecutePlugin(c *ConsumerFile,
		w http.ResponseWriter, r *http.Request) bool
	Config() interface{}
	Type() string
}

type Endpoint struct {
	Namespace  string `json:"-"`
	FilePath   string `json:"file_path,omitempty"`
	Path       string `json:"path,omitempty"`
	ServerPath string `json:"server_path"`

	Methods        []string `json:"methods"`
	AllowAnonymous bool     `json:"allow_anonymous"`
	Timeout        int      `json:"timeout"`

	AuthPluginInstances     []PluginInstance `json:"-"`
	InboundPluginInstances  []PluginInstance `json:"-"`
	TargetPluginInstance    PluginInstance   `json:"-"`
	OutboundPluginInstances []PluginInstance `json:"-"`
	Errors                  []string         `json:"errors"`
	Warnings                []string         `json:"warnings"`

	Plugins Plugins `json:"plugins"`
}

func ParseConsumerFile(data []byte) (*ConsumerFile, error) {
	res := &ConsumerFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v1") {
		return nil, fmt.Errorf("invalid consumer api version")
	}

	return res, nil
}

func ParseEndpointFile(data []byte) (*EndpointFile, error) {
	res := &EndpointFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v1") {
		return nil, fmt.Errorf("invalid endpoint api version")
	}

	return res, nil
}
