package target

import (
	"context"
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

type TargetWorkflowVarConfig struct {
	Namespace   string `yaml:"namespace" mapstructure:"namespace"`
	Flow        string `yaml:"flow" mapstructure:"flow"`
	Variable    string `yaml:"variable" mapstructure:"variable"`
	ContentType string `yaml:"content_type"  mapstructure:"content_type"`
}

// TargetFlowVarPlugin returns a workflow variable
type TargetFlowVarPlugin struct {
	config *TargetWorkflowVarConfig
}

func ConfigureWorkflowVar(config interface{}, ns string) (plugins.PluginInstance, error) {
	targetflowVarConfig := &TargetWorkflowVarConfig{}

	err := plugins.ConvertConfig(TargetNamespaceFilePluginName, config, targetflowVarConfig)
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

	return &TargetFlowVarPlugin{
		config: targetflowVarConfig,
	}, nil
}

func (tfv TargetFlowVarPlugin) Config() interface{} {
	return tfv.config
}

func (tfv TargetFlowVarPlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {

	url, err := createURLFlowVar(tfv.config.Namespace, tfv.config.Flow,
		tfv.config.Variable)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create url", err)
		return false
	}

	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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
