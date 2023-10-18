package spec

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type EndpointFile struct {
	DirektivAPI string `yaml:"direktiv_api"`
	Path        string `yaml:"path"`
	Method      string `yaml:"method"`
	Targets     []struct {
		Method string `yaml:"method"`
		Host   string `yaml:"host"`
		Path   string `yaml:"path"`
		Scheme string `yaml:"scheme"`
	} `yaml:"targets"`
	TimeoutSeconds int `yaml:"timeout_seconds"`
	PluginsConfig  []struct {
		Name          string      `yaml:"name"`
		Version       string      `yaml:"version"`
		RuntimeConfig interface{} `yaml:"runtime_config"`
	} `yaml:"plugins_config"`
}

func ParsePluginRouteFile(data []byte) (*EndpointFile, error) {
	res := &EndpointFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v1") {
		return nil, errors.New("invalid service api version")
	}

	return res, nil
}
