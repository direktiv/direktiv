package inbound

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
	TargetFlowPluginName = "target-flow"
)

type TargetWorkflowConfig struct {
	Namespace string `yaml:"namespace"`
	Flow      string `yaml:"flow"`
	Async     bool   `yaml:"Async"`
}

// TargetFlowPlugin executes a flow in a configured namespace.
// Flows can be executed async and sync
type TargetFlowPlugin struct {
	config *TargetWorkflowConfig
}

func (tf TargetFlowPlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
	targetflowConfig := &TargetWorkflowConfig{
		Async: true,
	}

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

	url, err := createURL(tf.config.Namespace, tf.config.Flow, tf.config.Async)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create url", err)
		return false
	}

	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		url.String(), io.NopCloser(r.Body))
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create request", err)
		return false
	}
	defer r.Body.Close()

	resp, err := client.Do(req)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not execute flow", err)
		return false
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not serve response", err)
		return false
	}
	resp.Body.Close()

	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(TargetFlowPlugin{})
}

func createURL(ns, flow string, wait bool) (*url.URL, error) {
	ex := "execute"
	if wait {
		ex = "wait"
	}

	// prefix with slash if set relative
	if !strings.HasPrefix(flow, "/") {
		flow = "/" + flow
	}

	urlString := fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s?op=%s&ref=latest",
		os.Getenv("DIREKTIV_API_V1_PORT"), ns, flow, ex)

	return url.Parse(urlString)
}
