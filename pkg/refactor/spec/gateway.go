package spec

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type EndpointFile struct {
	DirektivAPI          string       `yaml:"direktiv_api"`
	Path                 string       `yaml:"path"`
	Method               string       `yaml:"method"`
	TargetPlugin         PluginFile   `yaml:"target_plugin"`
	TimeoutSeconds       int          `yaml:"timeout_seconds"`
	AuthPluginsConfig    []PluginFile `yaml:"auth_plugins_config"`
	RequestPluginsConfig []PluginFile `yaml:"request_plugins_config"`
}

type PluginFile struct {
	Name          string      `yaml:"name"`
	Version       string      `yaml:"version"`
	RuntimeConfig interface{} `yaml:"runtime_config"`
}

func ParseEndpointFile(data []byte) (*EndpointFile, error) {
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
