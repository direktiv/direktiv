package target

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

type FlowVarPlugin struct {
	Namespace   string `mapstructure:"namespace"`
	Flow        string `mapstructure:"flow"`
	Variable    string `mapstructure:"variable"`
	ContentType string `mapstructure:"content_type"`
}

func (tnv *FlowVarPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &FlowVarPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	if pl.Variable == "" {
		return nil, fmt.Errorf("variable required")
	}

	return pl, nil
}

func (tnv *FlowVarPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	currentNS := gateway.ExtractContextEndpoint(r).Namespace
	if tnv.Namespace == "" {
		tnv.Namespace = currentNS
	}
	if tnv.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil, nil
	}

	uri := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/variables/?name=%s&workflowPath=%s&raw=true",
		os.Getenv("DIREKTIV_API_PORT"), tnv.Namespace, tnv.Variable, tnv.Flow)

	resp, err := doRequest(r, http.MethodGet, uri, nil)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "couldn't execute downstream request")
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		gateway.WriteInternalError(r, w, nil, "none ok status downstream request: "+resp.Status)
		return nil, nil
	}
	defer resp.Body.Close()

	// copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if tnv.ContentType != "" {
		w.Header().Set("Content-Type", tnv.ContentType)
	}
	// copy the status code
	w.WriteHeader(resp.StatusCode)

	// copy the response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		gateway.WriteInternalError(r, w, nil, "couldn't write downstream response")
		return nil, nil
	}

	return w, r
}

func (tnv *FlowVarPlugin) Type() string {
	return "target-flow-var"
}

func init() {
	gateway.RegisterPlugin(&FlowVarPlugin{})
}
