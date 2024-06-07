package target

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
	"go.opentelemetry.io/otel/trace"
)

// FlowPlugin executes a flow in a configured namespace.
type FlowPlugin struct {
	Namespace   string `mapstructure:"namespace"`
	Flow        string `mapstructure:"flow"`
	Async       bool   `mapstructure:"async"`
	ContentType string `mapstructure:"content_type"`

	internalAsync string
}

func (tf *FlowPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &FlowPlugin{
		Async: false,
	}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	if pl.Flow == "" {
		return nil, fmt.Errorf("flow required")
	}

	if !strings.HasPrefix(pl.Flow, "/") {
		pl.Flow = "/" + pl.Flow
	}

	pl.internalAsync = "wait"
	if pl.Async {
		pl.internalAsync = "execute"
	}

	return pl, nil
}

func (tf *FlowPlugin) Type() string {
	return "target-flow"
}

func (tf *FlowPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	currentNS := gateway.ExtractContextEndpoint(r).Namespace
	if tf.Namespace == "" {
		tf.Namespace = currentNS
	}
	if tf.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil
	}

	tracer := trace.SpanFromContext(r.Context()).TracerProvider().Tracer("direktiv/flow")
	ctx, childSpan := tracer.Start(r.Context(), "target-workflow-plugin")
	defer childSpan.End()

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/instances?path=%s&wait=%s",
		os.Getenv("DIREKTIV_API_PORT"),
		tf.Namespace, url.QueryEscape(tf.Flow),
		fmt.Sprintf("%v", tf.internalAsync == "wait"))

	resp, err := doRequest(r.WithContext(ctx), http.MethodPost, url, r.Body)
	if err != nil {
		gateway.WriteForbiddenError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	defer resp.Body.Close()

	// Copy headers.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if tf.ContentType != "" {
		w.Header().Set("Content-Type", tf.ContentType)
	}

	// Flow engine always return 200 and sets the error information in the headers, so we need to process them.
	errorCode := resp.Header.Get("Direktiv-Instance-Error-Code")
	errorMessage := resp.Header.Get("Direktiv-Instance-Error-Message")
	instance := resp.Header.Get("Direktiv-Instance-Id")

	if errorCode != "" {
		gateway.WriteInternalError(r, w,
			fmt.Errorf("errCode: %s, errMessage: %s, instanceId: %s", errorCode, errorMessage, instance),
			"workflow run failed")

		return nil
	}

	// Copy the status code.
	w.WriteHeader(resp.StatusCode)

	// Copy the response body.
	if _, err := io.Copy(w, resp.Body); err != nil {
		gateway.WriteInternalError(r, w, nil, "couldn't write downstream response")
		return nil
	}

	return r
}

func init() {
	gateway.RegisterPlugin(&FlowPlugin{})
}
