package inbound

import (
	"context"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/pkg/errors"

	"github.com/mitchellh/mapstructure"
)

const (
	InstantResponsePluginName = "instant-response"
)

// InstantResponseConfig configures the status code and response for InstantResponsePlugin.
type InstantResponseConfig struct {
	StatusCode    int    `yaml:"status_code"`
	StatusMessage string `yaml:"status_message"`
}

// InstantResponsePlugin responds instantly with the provided status code and message.
type InstantResponsePlugin struct {
	config *InstantResponseConfig
}

func (ir InstantResponsePlugin) Configure(config interface{}) (plugins.Plugin, error) {
	// var ok bool
	irConfig := &InstantResponseConfig{
		StatusCode:    http.StatusOK,
		StatusMessage: "This is the end!",
	}

	if config != nil {

		err := mapstructure.Decode(config, &irConfig)
		if err != nil {
			return nil, errors.Wrap(err, "configuration for instant-response invalid")
		}

		// irConfig, ok = config.(*InstantResponseConfig)
		// if !ok {
		// 	return nil, fmt.Errorf("configuration for instant-response invalid")
		// }
	}

	return &InstantResponsePlugin{
		config: irConfig,
	}, nil
}

func (ir InstantResponsePlugin) Name() string {
	return InstantResponsePluginName
}

func (ir InstantResponsePlugin) Type() plugins.PluginType {
	return plugins.InboundPluginType
}

func (ir InstantResponsePlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
	w http.ResponseWriter, r *http.Request) bool {

	if isJSON(ir.config.StatusMessage) {
		w.Header().Add("Content-Type", "application/json")
	}

	w.WriteHeader(ir.config.StatusCode)

	// nolint
	w.Write([]byte(ir.config.StatusMessage))

	return false
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(InstantResponsePlugin{})
}
