package inbound

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

type NameKeys struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HeaderManipulationPlugin struct {
	HeadersToAdd    []NameKeys `mapstructure:"headers_to_add"`
	HeadersToModify []NameKeys `mapstructure:"headers_to_modify"`
	HeadersToRemove []NameKeys `mapstructure:"headers_to_remove"`
}

func (hp *HeaderManipulationPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &HeaderManipulationPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (hp *HeaderManipulationPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	for a := range hp.HeadersToAdd {
		h := hp.HeadersToAdd[a]
		r.Header.Add(h.Name, h.Value)
	}

	for a := range hp.HeadersToModify {
		h := hp.HeadersToModify[a]
		r.Header.Set(h.Name, h.Value)
	}

	for a := range hp.HeadersToRemove {
		h := hp.HeadersToRemove[a]
		r.Header.Del(h.Name)
	}

	return r
}

func (hp *HeaderManipulationPlugin) Type() string {
	return "header-manipulation"
}

func init() {
	gateway.RegisterPlugin(&HeaderManipulationPlugin{})
}
