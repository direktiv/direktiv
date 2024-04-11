package target

import (
	"fmt"
	"io"
	"net/http"
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
	if targetflowConfig.Namespace != ns && ns != core.MagicalGatewayNamespace {
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
		flowArg:      tf.config.Flow,
		execArg:      tf.config.internalAsync,
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
	// direktivWorkflowRequest     direktivRequestType = "wf"
	direktivWorkflowVarRequest  direktivRequestType = "wfvar"
	direktivNamespaceVarRequest direktivRequestType = "nsvar"
	// direktivFileRequest         direktivRequestType = "file"

	namespaceArg = "ns"
	flowArg      = "flow"
	execArg      = "execute"
	varArg       = "variable"
	pathArg      = "path"
)

func doWorkflowRequest(args map[string]string, w http.ResponseWriter, r *http.Request) *http.Response {
	defer r.Body.Close()

	url := fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s?op=%s&ref=latest",
		os.Getenv("DIREKTIV_API_V1_PORT"), args[namespaceArg], args[flowArg], args[execArg])

	return doRequest(w, r, http.MethodPost, url, r.Body)
}
