package target

import (
	"context"
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
}

func ConfigureInstantResponse(config interface{}) (plugins.PluginInstance, error) {
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

func (ir *InstantResponsePlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {
	if plugins.IsJSON(ir.config.StatusMessage) {
		w.Header().Add("Content-Type", "application/json")
	}

	w.WriteHeader(ir.config.StatusCode)

	// nolint
	w.Write([]byte(ir.config.StatusMessage))

	return true
}

func (ir *InstantResponsePlugin) Config() interface{} {
	return ir.config
}

// InstantResponseConfig configures the status code and response for InstantResponsePlugin.
// type InstantResponseConfig struct {
// 	StatusCode    int    `yaml:"status_code" mapstructure:"status_code"`
// 	StatusMessage string `yaml:"status_message" mapstructure:"status_message"`
// }

// // InstantResponsePlugin responds instantly with the provided status code and message.
// type InstantResponsePlugin struct {
// 	plugins.PluginBase
// 	config *InstantResponseConfig
// }

// func (ir InstantResponsePlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
// 	irConfig := &InstantResponseConfig{
// 		StatusCode:    http.StatusOK,
// 		StatusMessage: "This is the end!",
// 	}

// 	if config != nil {
// 		err := mapstructure.Decode(config, &irConfig)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "configuration for instant-response invalid")
// 		}
// 	}

// 	return &InstantResponsePlugin{
// 		config: irConfig,
// 	}, nil

// }

// func (ir InstantResponsePlugin) Config() interface{} {
// 	return ir.config
// }

// // func (ir InstantResponsePlugin) Name() string {
// // 	return InstantResponsePluginName
// // }

// // func (ir InstantResponsePlugin) Type() plugins.PluginType {
// // 	return plugins.InboundPluginType
// // }

// func (ir InstantResponsePlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
// 	w http.ResponseWriter, r *http.Request) bool {

// 	if plugins.IsJSON(ir.config.StatusMessage) {
// 		w.Header().Add("Content-Type", "application/json")
// 	}

// 	w.WriteHeader(ir.config.StatusCode)

// 	// nolint
// 	w.Write([]byte(ir.config.StatusMessage))

// 	return false
// }

// //nolint:gochecknoinits
// func init() {
// 	plugins.AddPluginToRegistry(InstantResponsePltype InstantResponseConfig struct {
// 	StatusCode    int    `yaml:"status_code" mapstructure:"status_code"`
// 	StatusMessage string `yaml:"status_message" mapstructure:"status_message"`
// }

// // InstantResponsePlugin responds instantly with the provided status code and message.
// type InstantResponsePlugin struct {
// 	plugins.PluginBase
// 	config *InstantResponseConfig
// }

// func (ir InstantResponsePlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
// 	irConfig := &InstantResponseConfig{
// 		StatusCode:    http.StatusOK,
// 		StatusMessage: "This is the end!",
// 	}

// 	if config != nil {
// 		err := mapstructure.Decode(config, &irConfig)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "configuration for instant-response invalid")
// 		}
// 	}

// 	return &InstantResponsePlugin{
// 		config: irConfig,
// 	}, nil

// }

// func (ir InstantResponsePlugin) Config() interface{} {
// 	return ir.config
// }

// // func (ir InstantResponsePlugin) Name() string {
// // 	return InstantResponsePluginName
// // }

// // func (ir InstantResponsePlugin) Type() plugins.PluginType {
// // 	return plugins.InboundPluginType
// // }

// func (ir InstantResponsePlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
// 	w http.ResponseWriter, r *http.Request) bool {

// 	if plugins.IsJSON(ir.config.StatusMessage) {
// 		w.Header().Add("Content-Type", "application/json")
// 	}

// 	w.WriteHeader(ir.config.StatusCode)

// 	// nolint
// 	w.Write([]byte(ir.config.StatusMessage))

// 	return false
// }

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		InstantResponsePluginName,
		plugins.TargetPluginType,
		ConfigureInstantResponse))
}
