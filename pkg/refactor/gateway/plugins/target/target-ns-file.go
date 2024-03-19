package target

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
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
	if targetNamespaceFileConfig.Namespace != ns && ns != core.MagicalGatewayNamespace {
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
	resp := doDirektivRequest(direktivFileRequest, map[string]string{
		namespaceArg: tnf.config.Namespace,
		pathArg:      tnf.config.File,
	}, w, r)
	if resp == nil {
		return false
	}
	defer resp.Body.Close()

	data, err := fetchObjectData(resp)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not fetch file data", err)

		return false
	}

	r.Header.Set("Content-Type", "application/unknown")

	if tnf.config.ContentType != "" {
		w.Header().Set("Content-Type", tnf.config.ContentType)
	} else {
		// nolint
		kind, _ := filetype.Match(data)
		if kind != filetype.Unknown {
			w.Header().Set("Content-Type", kind.MIME.Value)
		}
	}

	// w.Header().Set("Content-Type", mtype.String())
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// nolint
	w.Write(data)

	return true
}

// nolint
type Node struct {
	Namespace string `json:"namespace"`
	Node      struct {
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
	} `json:"node"`
	EventLogging string `json:"eventLogging"`
	Oid          string `json:"oid"`
	Source       string `json:"source"`
}

func fetchObjectData(res *http.Response) ([]byte, error) {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var node Node
	err = json.Unmarshal(b, &node)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(node.Source)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		NamespaceFilePluginName,
		plugins.TargetPluginType,
		ConfigureNamespaceFilePlugin))
}
