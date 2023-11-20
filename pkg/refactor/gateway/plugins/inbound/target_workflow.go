package inbound

import (
	"context"
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	TargetFlowPluginName = "target-flow"
)

// RequestConvertConfig converts the whole request into JSON.
type TargetWorkflowConfig struct {
	Namespace string `yaml:"namespace"`
	Flow      string `yaml:"flow"`
}

// RequestConvertPlugin converts headers, query parameters, url paramneters
// and the body into a JSON object. The original body is discarded.
type TargetFlowPlugin struct {
	config *TargetWorkflowConfig
}

func (tf TargetFlowPlugin) Configure(config interface{}) (plugins.Plugin, error) {
	targetflowConfig := &TargetWorkflowConfig{}

	if config != nil {
		err := mapstructure.Decode(config, &targetflowConfig)
		if err != nil {
			return nil, errors.Wrap(err, "configuration for target-flow invalid")
		}
	}

	// set default to gateway namespace
	if targetflowConfig.Namespace == "" {
		targetflowConfig.Namespace = core.MagicalGatewayNamespace
	}

	return &TargetFlowPlugin{
		config: targetflowConfig,
	}, nil
}

func (tf TargetFlowPlugin) Name() string {
	return TargetFlowPluginName
}

func (tf TargetFlowPlugin) Type() plugins.PluginType {
	return plugins.InboundPluginType
}

func (tf TargetFlowPlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
	w http.ResponseWriter, r *http.Request) bool {

	fmt.Println("!!!!!")
	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(TargetFlowPlugin{})
}
