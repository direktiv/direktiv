package target

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

const (
	TargetFlowVarPluginName = "target-flow-var"
)

type WorkflowVarConfig struct {
	Namespace   string `mapstructure:"namespace"    yaml:"namespace"`
	Flow        string `mapstructure:"flow"         yaml:"flow"`
	Variable    string `mapstructure:"variable"     yaml:"variable"`
	ContentType string `mapstructure:"content_type" yaml:"content_type"`
}

// TargetFlowVarPlugin returns a workflow variable.
type FlowVarPlugin struct {
	config *WorkflowVarConfig
}

func ConfigureWorkflowVar(config interface{}, ns string) (core.PluginInstance, error) {
	targetflowVarConfig := &WorkflowVarConfig{}

	err := plugins.ConvertConfig(config, targetflowVarConfig)
	if err != nil {
		return nil, err
	}

	if targetflowVarConfig.Flow == "" || targetflowVarConfig.Variable == "" {
		return nil, fmt.Errorf("flow and variable required")
	}

	// set default to gateway namespace
	if targetflowVarConfig.Namespace == "" {
		targetflowVarConfig.Namespace = ns
	}

	// throw error if non magic namespace targets different namespace
	if targetflowVarConfig.Namespace != ns && ns != core.SystemNamespace {
		return nil, fmt.Errorf("plugin can not target different namespace")
	}

	if !strings.HasPrefix(targetflowVarConfig.Flow, "/") {
		targetflowVarConfig.Flow = "/" + targetflowVarConfig.Flow
	}

	return &FlowVarPlugin{
		config: targetflowVarConfig,
	}, nil
}

func (tfv FlowVarPlugin) Config() interface{} {
	return tfv.config
}

func (tfv FlowVarPlugin) ExecutePlugin(_ *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	// request failed if nil and response already written
	resp := doVariableRequest(direktivWorkflowVarRequest, map[string]string{
		namespaceArg: tfv.config.Namespace,
		pathArg:      tfv.config.Flow,
		varArg:       tfv.config.Variable,
	}, w, r)
	if resp == nil {
		return false
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not fetch file data", err)

		return false
	}

	var node Node
	err = json.Unmarshal(b, &node)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not fetch file data", err)

		return false
	}

	data := node.Data.Data

	// set headers from Direktiv
	w.Header().Set("Content-Type", node.Data.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(data)))

	// overwrite content type
	if tfv.config.ContentType != "" {
		w.Header().Set("Content-Type", tfv.config.ContentType)
	}

	// nolint
	w.Write(data)

	return true
}

func (tfv FlowVarPlugin) Type() string {
	return TargetFlowVarPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		TargetFlowVarPluginName,
		plugins.TargetPluginType,
		ConfigureWorkflowVar))
}
