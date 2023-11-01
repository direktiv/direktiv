package gateway

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type examplePlugin struct {
	conf examplePluginConfig
}

type examplePluginConfig struct {
	EchoValue string `json:"echo_value" jsonschema:"required"`
}

func (e examplePlugin) build(c map[string]interface{}) (serve, error) {
	var conf examplePluginConfig

	if err := unmarshalConfig(c, &conf); err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		slog.Info(conf.EchoValue)

		return true
	}, nil
}

func (e examplePlugin) getSchema() interface{} {
	return &e.conf
}

//nolint:gochecknoinits
func init() {
	registry["example_plugin"] = examplePlugin{}
}

func unmarshalConfig(c map[string]interface{}, target interface{}) error {
	// Convert the map to JSON bytes
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// Unmarshal the JSON bytes into the desired struct
	err = json.Unmarshal(jsonBytes, target)

	return err
}
