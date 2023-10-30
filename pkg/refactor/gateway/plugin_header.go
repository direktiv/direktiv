package gateway

import (
	"net/http"
)

type headerManipulationPlugin struct {
	HeadersToAdd    map[string]string `json:"headers_to_add"`
	HeadersToModify map[string]string `json:"headers_to_modify"`
	HeadersToRemove []string          `json:"headers_to_remove"`
}

type headerManipulationPluginConfig struct {
	HeadersToAdd    map[string]string `json:"headers_to_add"`
	HeadersToModify map[string]string `json:"headers_to_modify"`
	HeadersToRemove []string          `json:"headers_to_remove"`
}

func (p headerManipulationPlugin) build(c map[string]interface{}) (serve, error) {
	var conf headerManipulationPluginConfig

	if err := unmarshalConfig(c, &conf); err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		for key, value := range conf.HeadersToAdd {
			w.Header().Add(key, value)
		}

		for key, value := range conf.HeadersToModify {
			w.Header().Set(key, value)
		}

		for _, key := range conf.HeadersToRemove {
			w.Header().Del(key)
		}

		return true
	}, nil
}

func (p headerManipulationPlugin) getSchema() interface{} {
	return &headerManipulationPluginConfig{}
}

//nolint:gochecknoinits
func init() {
	registry["header_manipulation_plugin"] = headerManipulationPlugin{}
}
