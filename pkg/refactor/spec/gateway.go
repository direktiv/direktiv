package spec

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type PluginRouteFile struct {
	DirektivAPI string
	Path        string
	Method      string
	Targets     []struct {
		Method string
		Host   string
		Path   string
		Scheme string
	}
	TimeoutSeconds int
	PluginsConfig  []struct {
		Name                    string
		Version                 string
		Comment                 string
		Type                    string
		Priority                int               `yaml:"priority"`
		ExecutionTimeoutSeconds int               `yaml:"execution_timeout_seconds"`
		RuntimeConfig           map[string]string `yaml:"runtime_config"`
	}
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

type PluginRouteServiceDefinition struct {
}
