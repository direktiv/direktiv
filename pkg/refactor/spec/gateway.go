package spec

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type EndpointFile struct {
	DirektivAPI    string         `yaml:"direktiv_api"`
	Methods        []string       `yaml:"methods"`
	PathExtension  string         `yaml:"path_extension"`
	AllowAnonymous bool           `yaml:"allow_anonymous"`
	Plugins        []PluginConfig `yaml:"plugins"`
}

type ConsumerFile struct {
	DirektivAPI string   `yaml:"direktiv_api"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	APIKey      string   `yaml:"api_key"`
	Tags        []string `yaml:"tags"`
	Groups      []string `yaml:"groups"`
}

type PluginConfig struct {
	Type          string      `yaml:"type"`
	Configuration interface{} `yaml:"configuration"`
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
