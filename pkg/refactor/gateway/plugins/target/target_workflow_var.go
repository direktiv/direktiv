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
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	TargetFlowVarPluginName = "target-flow-var"
)

type TargetWorkflowVarConfig struct {
	Namespace string `yaml:"namespace"`
	Flow      string `yaml:"flow"`
	Variable  string `yaml:"variable"`
}

// TargetFlowVarPlugin returns a workflow variable
type TargetFlowVarPlugin struct {
	config *TargetWorkflowVarConfig
}

func (tfv TargetFlowVarPlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
	targetflowVarConfig := &TargetWorkflowVarConfig{}

	if config != nil {
		err := mapstructure.Decode(config, &targetflowVarConfig)
		if err != nil {
			return nil, errors.Wrap(err, "configuration for target-flow-var invalid")
		}
	}

	// set default to gateway namespace
	if targetflowVarConfig.Namespace == "" {
		targetflowVarConfig.Namespace = core.MagicalGatewayNamespace
	}

	return &TargetFlowVarPlugin{
		config: targetflowVarConfig,
	}, nil
}

func (tfv TargetFlowVarPlugin) Config() interface{} {
	return tfv.config
}

func (tfv TargetFlowVarPlugin) Name() string {
	return TargetFlowVarPluginName
}

func (tfv TargetFlowVarPlugin) Type() plugins.PluginType {
	return plugins.InboundPluginType
}

func (tfv TargetFlowVarPlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
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

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not serve variable", err)
		return false
	}
	resp.Body.Close()

	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(TargetFlowVarPlugin{})
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
