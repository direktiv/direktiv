package spec

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type PluginRouteFile struct {
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
		Name                    string            `yaml:"name"`
		Version                 string            `yaml:"version"`
		Comment                 string            `yaml:"comment"`
		Type                    string            `yaml:"type"`
		Priority                int               `yaml:"priority"`
		ExecutionTimeoutSeconds int               `yaml:"execution_timeout_seconds"`
		RuntimeConfig           map[string]string `yaml:"runtime_config"`
	} `yaml:"plugins_config"`
}

func ParsePluginRouteFile(data []byte) (*PluginRouteFile, error) {
	res := &PluginRouteFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "pluginroute/v1") {
		return nil, errors.New("invalid service api version")
	}

	return res, nil
}
