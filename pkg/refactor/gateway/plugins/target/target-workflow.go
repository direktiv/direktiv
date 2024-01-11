package target

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/util"
	"go.opentelemetry.io/otel/trace"
)

const (
	FlowPluginName = "target-flow"
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
	resp := doDirektivRequest(direktivWorkflowRequest, map[string]string{
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
		plugins.ReportError(w, http.StatusInternalServerError,
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
	direktivWorkflowRequest     direktivRequestType = "wf"
	direktivWorkflowVarRequest  direktivRequestType = "wfvar"
	direktivNamespaceVarRequest direktivRequestType = "nsvar"
	direktivFileRequest         direktivRequestType = "file"

	namespaceArg = "ns"
	flowArg      = "flow"
	execArg      = "execute"
	varArg       = "variable"
	pathArg      = "path"
)

type httpCarrier struct {
	r *http.Request
}

func (c *httpCarrier) Get(key string) string {
	return c.r.Header.Get(key)
}

func (c *httpCarrier) Keys() []string {
	return c.r.Header.Values("oteltmckeys")
}

func (c *httpCarrier) Set(key, val string) {
	prev := c.Get(key)
	if prev == "" {
		c.r.Header.Add("oteltmckeys", key)
	}
	c.r.Header.Set(key, val)
}

func doDirektivRequest(requestType direktivRequestType, args map[string]string,
	w http.ResponseWriter, r *http.Request,
) *http.Response {
	defer r.Body.Close()

	var (
		url    string
		method = http.MethodGet
		body   io.ReadCloser
	)

	switch requestType {
	case direktivFileRequest:
		url = fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s",
			os.Getenv("DIREKTIV_API_V1_PORT"), args[namespaceArg], args[pathArg])
	case direktivWorkflowVarRequest:
		url = fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s?op=var&var=%s",
			os.Getenv("DIREKTIV_API_V1_PORT"), args[namespaceArg], args[flowArg], args[varArg])
	case direktivNamespaceVarRequest:
		url = fmt.Sprintf("http://localhost:%s/api/namespaces/%s/vars/%s",
			os.Getenv("DIREKTIV_API_V1_PORT"), args[namespaceArg], args[varArg])
	case direktivWorkflowRequest:
		fallthrough
	default:
		// workflow request is default
		method = http.MethodPost
		body = r.Body
		url = fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s?op=%s&ref=latest",
			os.Getenv("DIREKTIV_API_V1_PORT"), args[namespaceArg], args[flowArg], args[execArg])
	}

	client := http.Client{}
	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, method, url, body)

	endTrace := util.TraceGWHTTPRequest(ctx, req, "direktiv/flow")
	defer endTrace()
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create request", err)

		return nil
	}

	// add api key if required
	if os.Getenv("DIREKTIV_API_KEY") != "" {
		req.Header.Set("Direktiv-Token", os.Getenv("DIREKTIV_API_KEY"))
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not execute flow", err)

		return nil
	}

	// error handling
	errorCode := resp.Header.Get("Direktiv-Instance-Error-Code")
	errorMessage := resp.Header.Get("Direktiv-Instance-Error-Message")
	instance := resp.Header.Get("Direktiv-Instance-Id")

	if errorCode != "" {
		msg := fmt.Sprintf("%s: %s (%s)", errorCode, errorMessage, instance)
		plugins.ReportError(w, resp.StatusCode,
			"error executing workflow", fmt.Errorf(msg))

		return nil
	}

	// direktiv requests always respond with 200, workflow errors are handled in the previous check
	if resp.StatusCode >= http.StatusMultipleChoices {
		plugins.ReportError(w, resp.StatusCode,
			"can not execute flow", fmt.Errorf(resp.Status))

		return nil
	}

	return resp
}
