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
	"go.opentelemetry.io/otel/trace"
)

const (
	FlowPluginName     = "target-flow"
	defaultContentType = "application/json"
)

type WorkflowConfig struct {
	Namespace   string `mapstructure:"namespace"    yaml:"namespace"`
	Flow        string `mapstructure:"flow"         yaml:"flow"`
	Async       bool   `mapstructure:"async"        yaml:"async"`
	ContentType string `mapstructure:"content_type" yaml:"content_type"`

	internalAsync string
}

// TargetFlowPlugin executes a flow in a configured namespace.
// Flows can be executed async and sync.
type FlowPlugin struct {
	config *WorkflowConfig
}

func ConfigureTargetFlowPlugin(config interface{}, ns string) (core.PluginInstance, error) {
	targetflowConfig := &WorkflowConfig{
		Async: false,
	}

	err := plugins.ConvertConfig(config, targetflowConfig)
	if err != nil {
		return nil, err
	}

	if targetflowConfig.Flow == "" {
		return nil, fmt.Errorf("flow required")
	}

	// set default to gateway namespace
	if targetflowConfig.Namespace == "" {
		targetflowConfig.Namespace = ns
	}

	// if content type is not set use application/json
	if targetflowConfig.ContentType == "" {
		targetflowConfig.ContentType = defaultContentType
	}

	// throw error if non magic namespace targets different namespace
	if targetflowConfig.Namespace != ns && ns != core.SystemNamespace {
		return nil, fmt.Errorf("plugin can not target different namespace")
	}

	if !strings.HasPrefix(targetflowConfig.Flow, "/") {
		targetflowConfig.Flow = "/" + targetflowConfig.Flow
	}

	targetflowConfig.internalAsync = "wait"
	if targetflowConfig.Async {
		targetflowConfig.internalAsync = "execute"
	}

	return &FlowPlugin{
		config: targetflowConfig,
	}, nil
}

func (tf FlowPlugin) Type() string {
	return FlowPluginName
}

func (tf FlowPlugin) Config() interface{} {
	return tf.config
}

func (tf FlowPlugin) ExecutePlugin(_ *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	tracer := trace.SpanFromContext(r.Context()).TracerProvider().Tracer("direktiv/flow")
	ctx, childSpan := tracer.Start(r.Context(), "target-workflow-plugin")
	defer childSpan.End()
	// request failed if nil and response already written
	resp := doWorkflowRequest(map[string]string{
		namespaceArg: tf.config.Namespace,
		flowArg:      url.QueryEscape(tf.config.Flow),
		execArg:      fmt.Sprintf("%v", tf.config.internalAsync == "wait"),
	}, w, r.WithContext(ctx))
	if resp == nil {
		return false
	}

	if tf.config.ContentType != "" {
		w.Header().Set("Content-Type", tf.config.ContentType)
	}

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not serve response", err)

		return false
	}
	resp.Body.Close()

	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		FlowPluginName,
		plugins.TargetPluginType,
		ConfigureTargetFlowPlugin))
}

type direktivRequestType string

const (
	direktivWorkflowVarRequest  direktivRequestType = "wfvar"
	direktivNamespaceVarRequest direktivRequestType = "nsvar"

	namespaceArg = "ns"
	flowArg      = "flow"
	execArg      = "execute"
	varArg       = "variable"
	pathArg      = "path"
)

func doWorkflowRequest(args map[string]string, w http.ResponseWriter, r *http.Request) *http.Response {
	defer r.Body.Close()

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/instances?path=%s&wait=%s",
		os.Getenv("DIREKTIV_API_PORT"), args[namespaceArg], args[flowArg], args[execArg])

	resp := doRequest(w, r, http.MethodPost, url, r.Body)
	if resp == nil {
		return nil
	}

	// error handling
	errorCode := resp.Header.Get("Direktiv-Instance-Error-Code")
	errorMessage := resp.Header.Get("Direktiv-Instance-Error-Message")
	instance := resp.Header.Get("Direktiv-Instance-Id")

	if errorCode != "" {
		msg := fmt.Sprintf("%s: %s (%s)", errorCode, errorMessage, instance)
		plugins.ReportError(r.Context(), w, resp.StatusCode,
			"error executing workflow", fmt.Errorf(msg))

		return nil
	}

	// direktiv requests always respond with 200, workflow errors are handled in the previous check
	if resp.StatusCode >= http.StatusMultipleChoices {
		plugins.ReportError(r.Context(), w, resp.StatusCode,
			"can not execute flow", fmt.Errorf(resp.Status))

		return nil
	}

	return resp
}
