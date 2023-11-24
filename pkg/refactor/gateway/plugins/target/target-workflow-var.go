package target

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
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

func ConfigureWorkflowVar(config interface{}, ns string) (plugins.PluginInstance, error) {
	targetflowVarConfig := &WorkflowVarConfig{}

	err := plugins.ConvertConfig(config, targetflowVarConfig)
	if err != nil {
		return nil, err
	}

	// set default to gateway namespace
	if targetflowVarConfig.Namespace == "" {
		targetflowVarConfig.Namespace = ns
	}

	// throw error if non magic namespace targets different namespace
	if targetflowVarConfig.Namespace != ns && ns != core.MagicalGatewayNamespace {
		return nil, fmt.Errorf("plugin can not target different namespace")
	}

	return &FlowVarPlugin{
		config: targetflowVarConfig,
	}, nil
}

func (tfv FlowVarPlugin) Config() interface{} {
	return tfv.config
}

func (tfv FlowVarPlugin) ExecutePlugin(_ *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	url, err := createURLFlowVar(tfv.config.Namespace, tfv.config.Flow,
		tfv.config.Variable)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create url", err)

		return false
	}

	client := http.Client{}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url.String(), nil)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create request", err)

		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not serve variable", err)

		return false
	}

	// set headers from Direktiv
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))

	// overwrite content type
	if tfv.config.ContentType != "" {
		w.Header().Set("Content-Type", tfv.config.ContentType)
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not serve variable", err)

		return false
	}
	resp.Body.Close()

	return true
}

func createURLFlowVar(ns, flow, v string) (*url.URL, error) {
	// prefix with slash if set relative
	if !strings.HasPrefix(flow, "/") {
		flow = "/" + flow
	}

	urlString := fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s?op=var&var=%s",
		os.Getenv("DIREKTIV_API_V1_PORT"), ns, flow, v)

	return url.Parse(urlString)
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		TargetFlowVarPluginName,
		plugins.TargetPluginType,
		ConfigureWorkflowVar))
}
