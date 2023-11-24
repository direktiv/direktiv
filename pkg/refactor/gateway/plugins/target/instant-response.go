package target

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
)

const (
	InstantResponsePluginName = "instant-response"
)

type InstantResponsePlugin struct {
	config *InstantResponseConfig
}

type InstantResponseConfig struct {
	StatusCode    int    `yaml:"status_code" mapstructure:"status_code"`
	StatusMessage string `yaml:"status_message" mapstructure:"status_message"`
	ContentType   string `yaml:"content_type"  mapstructure:"content_type"`
}

func ConfigureInstantResponse(config interface{}, ns string) (plugins.PluginInstance, error) {
	irConfig := &InstantResponseConfig{
		StatusCode:    http.StatusOK,
		StatusMessage: "This is the end!",
	}

	err := plugins.ConvertConfig(InstantResponsePluginName, config, irConfig)
	if err != nil {
		return nil, err
	}

	return &InstantResponsePlugin{
		config: irConfig,
	}, nil
}

func (ir *InstantResponsePlugin) ExecutePlugin(c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {
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
