package target

import (
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultContentType = "application/json"
)

// FlowPlugin executes a flow in a configured namespace.
type FlowPlugin struct {
	Namespace string `mapstructure:"namespace"`
	Flow      string `mapstructure:"flow"`
	Async     bool   `mapstructure:"async"`
	// TODO: yassir, need fix here.
	//ContentType string `mapstructure:"content_type"`

	internalAsync string
}

func (tf *FlowPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &FlowPlugin{
		Async: false,
	}

	err := gateway2.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	if pl.Flow == "" {
		return nil, fmt.Errorf("flow required")
	}

	// TODO: yassir, need fix here.
	// if content type is not set use application/json
	//if pl.ContentType == "" {
	//	pl.ContentType = defaultContentType
	//}

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
	currentNS := gateway2.ExtractContextEndpoint(r).Namespace
	if tf.Namespace == "" {
		tf.Namespace = currentNS
	}
	if tf.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway2.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil
	}

	tracer := trace.SpanFromContext(r.Context()).TracerProvider().Tracer("direktiv/flow")
	ctx, childSpan := tracer.Start(r.Context(), "target-workflow-plugin")
	defer childSpan.End()

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/instances?path=%s&wait=%s",
		os.Getenv("DIREKTIV_API_V1_PORT"),
		tf.Namespace, url.QueryEscape(tf.Flow),
		fmt.Sprintf("%v", tf.internalAsync == "wait"))

	resp, err := doRequest(r.WithContext(ctx), http.MethodPost, url, r.Body)
	if err != nil {
		gateway2.WriteForbiddenError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	defer resp.Body.Close()

	// TODO: yassir, need fix here.
	//if tf.ContentType != "" {
	//	w.Header().Set("Content-Type", tf.ContentType)
	//}
	//errorCode := resp.Header.Get("Direktiv-Instance-Error-Code")
	//errorMessage := resp.Header.Get("Direktiv-Instance-Error-Message")
	//instance := resp.Header.Get("Direktiv-Instance-Id")
	//
	//if errorCode != "" {
	//	msg := fmt.Sprintf("%s: %s (%s)", errorCode, errorMessage, instance)
	//	plugins.ReportError(r.Context(), w, resp.StatusCode,
	//		"error executing workflow", fmt.Errorf(msg))
	//
	//	return nil
	//}
	//
	//// direktiv requests always respond with 200, workflow errors are handled in the previous check
	//if resp.StatusCode >= http.StatusMultipleChoices {
	//	plugins.ReportError(r.Context(), w, resp.StatusCode,
	//		"can not execute flow", fmt.Errorf(resp.Status))
	//
	//	return nil
	//}

	// copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	// copy the status code
	w.WriteHeader(resp.StatusCode)

	// copy the response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't write downstream response")
		return nil
	}

	return r
}

func init() {
	gateway2.RegisterPlugin(&FlowPlugin{})
}
