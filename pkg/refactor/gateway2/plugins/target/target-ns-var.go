package target

import (
	"encoding/json"
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"io"
	"net/http"
	"os"
)

type NamespaceVarPlugin struct {
	Namespace string `mapstructure:"namespace"`
	Variable  string `mapstructure:"variable"`
}

func (tnv *NamespaceVarPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &NamespaceVarPlugin{}

	err := gateway2.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	if pl.Variable == "" {
		return nil, fmt.Errorf("variable required")
	}

	return pl, nil
}

func (tnv *NamespaceVarPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	currentNS := gateway2.ExtractContextEndpoint(r).Namespace
	if tnv.Namespace == "" {
		tnv.Namespace = currentNS
	}
	if tnv.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway2.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil
	}

	uri := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/variables/?name=%s",
		os.Getenv("DIREKTIV_API_V2_PORT"), tnv.Namespace, tnv.Variable)

	resp, err := doRequest(r, http.MethodGet, uri, nil)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request1")
		return nil
	}
	defer resp.Body.Close()

	type object struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	obj := &object{}
	err = json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	if len(obj.Data) > 0 {
		uri = fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/variables/%s",
			os.Getenv("DIREKTIV_API_V2_PORT"), tnv.Namespace, obj.Data[0].ID)
		resp, err = doRequest(r, http.MethodGet, uri, nil)
		if err != nil {
			gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request")
			return nil
		}
		defer resp.Body.Close()
	}

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

func (tnv *NamespaceVarPlugin) Type() string {
	return "target-namespace-var"
}

func init() {
	gateway2.RegisterPlugin(&NamespaceVarPlugin{})
}
