package target

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

const (
	InstantResponsePluginName = "instant-response"
)

type InstantResponsePlugin struct {
	config *InstantResponseConfig
}

type InstantResponseConfig struct {
	StatusCode    int    `mapstructure:"status_code"    yaml:"status_code"`
	StatusMessage string `mapstructure:"status_message" yaml:"status_message"`
	ContentType   string `mapstructure:"content_type"   yaml:"content_type"`
}

func ConfigureInstantResponse(config interface{}, _ string) (plugins.PluginInstance, error) {
	irConfig := &InstantResponseConfig{
		StatusCode:    http.StatusOK,
		StatusMessage: "This is the end!",
	}

	err := plugins.ConvertConfig(config, irConfig)
	if err != nil {
		return nil, err
	}

	return &InstantResponsePlugin{
		config: irConfig,
	}, nil
}

func (ir *InstantResponsePlugin) ExecutePlugin(_ *core.ConsumerBase,
	w http.ResponseWriter, _ *http.Request,
) bool {
	if plugins.IsJSON(ir.config.StatusMessage) {
		w.Header().Add("Content-Type", "application/json")
	}

	w.WriteHeader(ir.config.StatusCode)

	if ir.config.ContentType != "" {
		w.Header().Set("Content-Type", ir.config.ContentType)
	}

	// nolint
	w.Write([]byte(ir.config.StatusMessage))

	return true
}

func (ir *InstantResponsePlugin) Config() interface{} {
	return ir.config
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		InstantResponsePluginName,
		plugins.TargetPluginType,
		ConfigureInstantResponse))
}
