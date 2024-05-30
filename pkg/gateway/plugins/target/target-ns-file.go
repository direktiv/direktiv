package target

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	"github.com/h2non/filetype"
)

const (
	NamespaceFilePluginName = "target-namespace-file"
)

type NamespaceFileConfig struct {
	Namespace   string `mapstructure:"namespace"    yaml:"namespace"`
	File        string `mapstructure:"file"         yaml:"file"`
	ContentType string `mapstructure:"content_type" yaml:"content_type"`
}

// TargetNamespaceFilePlugin returns a files in the explorer tree.
type NamespaceFilePlugin struct {
	config *NamespaceFileConfig
}

func ConfigureNamespaceFilePlugin(config interface{}, ns string) (core.PluginInstance, error) {
	targetNamespaceFileConfig := &NamespaceFileConfig{}

	err := plugins.ConvertConfig(config, targetNamespaceFileConfig)
	if err != nil {
		return nil, err
	}

	if targetNamespaceFileConfig.File == "" {
		return nil, fmt.Errorf("file is required")
	}

	if !strings.HasPrefix(targetNamespaceFileConfig.File, "/") {
		targetNamespaceFileConfig.File = "/" + targetNamespaceFileConfig.File
	}

	// set default to gateway namespace
	if targetNamespaceFileConfig.Namespace == "" {
		targetNamespaceFileConfig.Namespace = ns
	}

	// throw error if non magic namespace targets different namespace
	if targetNamespaceFileConfig.Namespace != ns && ns != core.SystemNamespace {
		return nil, fmt.Errorf("plugin can not target different namespace")
	}

	return &NamespaceFilePlugin{
		config: targetNamespaceFileConfig,
	}, nil
}

func (tnf NamespaceFilePlugin) Config() interface{} {
	return tnf.config
}

func (tnf NamespaceFilePlugin) Type() string {
	return NamespaceFilePluginName
}

func (tnf NamespaceFilePlugin) ExecutePlugin(
	_ *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	// request failed if nil and response already written
	resp := doFilesystemRequest(map[string]string{
		namespaceArg: tnf.config.Namespace,
		pathArg:      tnf.config.File,
	}, w, r)
	if resp == nil {
		return false
	}
	defer resp.Body.Close()

	data, mime, err := fetchObjectData(resp)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not fetch file data", err)

		return false
	}

	mt := "application/unknown"

	// overwrite object mimetype if configured
	// otherwise use the one coming from the API
	// last resort is guessing
	if tnf.config.ContentType != "" {
		mt = tnf.config.ContentType
	} else if mime != "" {
		mt = mime
	} else {
		// guessing
		// nolint
		kind, _ := filetype.Match(data)
		if kind != filetype.Unknown {
			mt = kind.MIME.Value
		}
	}
	w.Header().Set("Content-Type", mt)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// nolint
	w.Write(data)

	return true
}

// nolint
type Node struct {
	Data struct {
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
		Name         string    `json:"name"`
		Path         string    `json:"path"`
		Parent       string    `json:"parent"`
		Type         string    `json:"type"`
		Attributes   []any     `json:"attributes"`
		Oid          string    `json:"oid"`
		ReadOnly     bool      `json:"readOnly"`
		ExpandedType string    `json:"expandedType"`
		MimeType     string    `json:"mimeType"`
		Data         []byte    `json:"data"`
	} `json:"data"`
}

func fetchObjectData(res *http.Response) ([]byte, string, error) {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}

	var node Node
	err = json.Unmarshal(b, &node)
	if err != nil {
		return nil, "", err
	}

	data := node.Data.Data

	return data, node.Data.MimeType, nil
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		NamespaceFilePluginName,
		plugins.TargetPluginType,
		ConfigureNamespaceFilePlugin))
}

func doFilesystemRequest(args map[string]string,
	w http.ResponseWriter, r *http.Request,
) *http.Response {
	defer r.Body.Close()

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s",
		os.Getenv("DIREKTIV_API_PORT"), args[namespaceArg], args[pathArg])

	return doRequest(w, r, http.MethodGet, url, nil)
}
