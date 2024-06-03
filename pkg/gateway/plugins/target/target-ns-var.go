package target

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	"github.com/google/uuid"
)

const (
	TargetNamespaceVarPluginName = "target-namespace-var"
)

type NamespaceVarConfig struct {
	Namespace   string `mapstructure:"namespace"    yaml:"namespace"`
	Variable    string `mapstructure:"variable"     yaml:"variable"`
	ContentType string `mapstructure:"content_type" yaml:"content_type"`
}

type NamespaceVarPlugin struct {
	config *NamespaceVarConfig
}

func ConfigureNamespaceVarPlugin(config interface{}, ns string) (core.PluginInstance, error) {
	targetNamespaceVarConfig := &NamespaceVarConfig{}

	err := plugins.ConvertConfig(config, targetNamespaceVarConfig)
	if err != nil {
		return nil, err
	}

	if targetNamespaceVarConfig.Variable == "" {
		return nil, fmt.Errorf("variable required")
	}

	// set default to gateway namespace
	if targetNamespaceVarConfig.Namespace == "" {
		targetNamespaceVarConfig.Namespace = ns
	}

	// throw error if non magic namespace targets different namespace
	if targetNamespaceVarConfig.Namespace != ns && ns != core.SystemNamespace {
		return nil, fmt.Errorf("plugin can not target different namespace")
	}

	return &NamespaceVarPlugin{
		config: targetNamespaceVarConfig,
	}, nil
}

func (tnv NamespaceVarPlugin) Config() interface{} {
	return tnv.config
}

func (tnv NamespaceVarPlugin) ExecutePlugin(_ *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	// request failed if nil and response already written
	resp := doVariableRequest(direktivNamespaceVarRequest, map[string]string{
		namespaceArg: tnv.config.Namespace,
		varArg:       tnv.config.Variable,
	}, w, r)
	if resp == nil {
		return false
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not fetch file data", err)

		return false
	}

	var node Node
	err = json.Unmarshal(b, &node)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not fetch file data", err)

		return false
	}

	data := node.Data.Data

	// set headers from Direktiv
	w.Header().Set("Content-Type", node.Data.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(data)))

	// overwrite content type
	if tnv.config.ContentType != "" {
		w.Header().Set("Content-Type", tnv.config.ContentType)
	}

	// nolint
	w.Write(data)

	return true
}

func (tnv NamespaceVarPlugin) Type() string {
	return TargetNamespaceVarPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		TargetNamespaceVarPluginName,
		plugins.TargetPluginType,
		ConfigureNamespaceVarPlugin))
}

type varResolveElem struct {
	ID        uuid.UUID `json:"id"`
	Typ       string    `json:"type"`
	Reference string    `json:"reference"`
	Name      string    `json:"name"`

	Size      int       `json:"size"`
	MimeType  string    `json:"mimeType"`
	Data      []byte    `json:"data,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type varResolveResponse struct {
	Data []varResolveElem
}

func doVariableRequest(requestType direktivRequestType, args map[string]string,
	w http.ResponseWriter, r *http.Request,
) *http.Response {
	defer r.Body.Close()

	varName := args[varArg]

	uri := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/variables/?name=%s",
		os.Getenv("DIREKTIV_API_PORT"), args[namespaceArg], varName)

	if requestType == direktivWorkflowVarRequest {
		uri = fmt.Sprintf("%s&workflowPath=%s", uri, url.QueryEscape(args[pathArg]))
	}

	resp := doRequest(w, r, http.MethodGet, uri, nil)
	if resp == nil {
		return nil
	}
	defer resp.Body.Close()

	expectedResponse := new(varResolveResponse)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"resolve request returned error", err)

		return nil
	}

	err = json.Unmarshal(data, &expectedResponse)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"resolve request returned unexpected response", err)

		return nil
	}

	if len(expectedResponse.Data) == 0 {
		plugins.ReportError(r.Context(), w, http.StatusNotFound,
			"variable not found", errors.New("not found"))

		return nil
	}

	varID := expectedResponse.Data[0].ID.String()

	uri = fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/variables/%s",
		os.Getenv("DIREKTIV_API_PORT"), args[namespaceArg], varID)

	resp = doRequest(w, r, http.MethodGet, uri, nil)
	if resp == nil {
		return nil
	}

	return resp
}
