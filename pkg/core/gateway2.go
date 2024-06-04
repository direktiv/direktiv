package core

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type GatewayManagerV2 interface {
	http.Handler

	SetEndpoints(list []EndpointV2, cList []ConsumerV2)
}

type EndpointFileV2 struct {
	DirektivAPI    string          `json:"-"                 yaml:"direktiv_api"`
	Methods        []string        `json:"methods,omitempty" yaml:"methods"`
	Path           string          `json:"path,omitempty"    yaml:"path"` // TODO: fix ui to remove server_path, warnings
	AllowAnonymous bool            `json:"allow_anonymous"   yaml:"allow_anonymous"`
	PluginsConfig  PluginsConfigV2 `json:"plugins,omitempty" yaml:"plugins"`
	Timeout        int             `json:"timeout"           yaml:"timeout"`
	Errors         []string        `json:"errors,omitempty"  yaml:"-"`
}

type ConsumerFileV2 struct {
	DirektivAPI string   `json:"-"        yaml:"direktiv_api"`
	Username    string   `json:"username" yaml:"username"`
	Password    string   `json:"password" yaml:"password"`
	APIKey      string   `json:"api_key"  yaml:"api_key"`
	Tags        []string `json:"tags"     yaml:"tags"`
	Groups      []string `json:"groups"   yaml:"groups"`
	Errors      []string `json:"errors"   yaml:"-"`
}

type PluginsConfigV2 struct {
	Auth     []PluginConfigV2 `json:"auth,omitempty"     yaml:"auth"`
	Inbound  []PluginConfigV2 `json:"inbound,omitempty"  yaml:"inbound"`
	Target   PluginConfigV2   `json:"target,omitempty"   yaml:"target"`
	Outbound []PluginConfigV2 `json:"outbound,omitempty" yaml:"outbound"`
}

type PluginConfigV2 struct {
	Typ    string         `json:"type,omitempty"          yaml:"type"`
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

	Namespace string `json:"-"         yaml:"-"`
	FilePath  string `json:"file_path" yaml:"-"`
}

type ConsumerV2 struct {
	ConsumerFileV2

	Namespace string `json:"-"         yaml:"-"`
	FilePath  string `json:"file_path" yaml:"-"`
}

func ParseConsumerFileV2(data []byte) (*ConsumerFileV2, error) {
	res := &ConsumerFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v1") {
		return nil, fmt.Errorf("invalid consumer api version")
	}

	return res, nil
}

func ParseEndpointFileV2(data []byte) (*EndpointFileV2, error) {
	res := &EndpointFileV2{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v1") {
		return nil, fmt.Errorf("invalid endpoint api version")
	}

	return res, nil
}
