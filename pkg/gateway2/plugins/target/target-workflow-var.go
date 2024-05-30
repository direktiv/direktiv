package target

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
)

type FlowVarPlugin struct {
	Namespace string `mapstructure:"namespace"`
	Flow      string `mapstructure:"flow"`
	Variable  string `mapstructure:"variable"`
}

func (tnv *FlowVarPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &FlowVarPlugin{}

	err := gateway2.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	if pl.Variable == "" {
		return nil, fmt.Errorf("variable required")
	}

	return pl, nil
}

func (tnv *FlowVarPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	currentNS := gateway2.ExtractContextEndpoint(r).Namespace
	if tnv.Namespace == "" {
		tnv.Namespace = currentNS
	}
	if tnv.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway2.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil
	}

	uri := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/variables/?name=%s&workflowPath=%s&raw=true",
		os.Getenv("DIREKTIV_API_PORT"), tnv.Namespace, tnv.Variable, tnv.Flow)

	resp, err := doRequest(r, http.MethodGet, uri, nil)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	defer resp.Body.Close()

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

func (tnv *FlowVarPlugin) Type() string {
	return "target-flow-var"
}

func init() {
	gateway2.RegisterPlugin(&FlowVarPlugin{})
}
