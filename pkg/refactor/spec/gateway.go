package spec

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type EndpointFile struct {
	DirektivAPI string         `yaml:"direktiv_api"`
	Method      string         `yaml:"method"`
	Workflow    string         `yaml:"workflow"`
	Namespace   string         `yaml:"namespace"`
	Plugins     []PluginConfig `yaml:"plugins"`
}

type PluginConfig struct {
	Type          string                 `yaml:"type"`
	Configuration map[string]interface{} `yaml:"configuration"`
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
