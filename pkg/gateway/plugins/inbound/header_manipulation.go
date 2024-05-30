package inbound

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
)

const (
	HeaderManipulation = "header-manipulation"
)

type NameKeys struct {
	Name  string `json:"name"  yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type HeaderManipulationConfig struct {
	HeadersToAdd    []NameKeys `json:"headers_to_add"    mapstructure:"headers_to_add"    yaml:"headers_to_add"`
	HeadersToModify []NameKeys `json:"headers_to_modify" mapstructure:"headers_to_modify" yaml:"headers_to_modify"`
	HeadersToRemove []NameKeys `json:"headers_to_remove" mapstructure:"headers_to_remove" yaml:"headers_to_remove"`
}

type HeaderManipulationPlugin struct {
	configuration *HeaderManipulationConfig
}

func ConfigureHeaderManipulation(config interface{}, _ string) (core.PluginInstance, error) {
	headerManipulationConfig := &HeaderManipulationConfig{}

	err := plugins.ConvertConfig(config, headerManipulationConfig)
	if err != nil {
		return nil, err
	}

	return &HeaderManipulationPlugin{
		configuration: headerManipulationConfig,
	}, nil
}

func (hp *HeaderManipulationPlugin) Config() interface{} {
	return hp.configuration
}

func (hp *HeaderManipulationPlugin) ExecutePlugin(_ *core.ConsumerFile,
	_ http.ResponseWriter, r *http.Request,
) bool {
	for a := range hp.configuration.HeadersToAdd {
		h := hp.configuration.HeadersToAdd[a]
		r.Header.Add(h.Name, h.Value)
	}

	for a := range hp.configuration.HeadersToModify {
		h := hp.configuration.HeadersToModify[a]
		r.Header.Set(h.Name, h.Value)
	}

	for a := range hp.configuration.HeadersToRemove {
		h := hp.configuration.HeadersToRemove[a]
		r.Header.Del(h.Name)
	}

	return true
}

func (hp *HeaderManipulationPlugin) Type() string {
	return HeaderManipulation
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		HeaderManipulation,
		plugins.InboundPluginType,
		ConfigureHeaderManipulation))
}
