package target

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"github.com/h2non/filetype"
)

const (
	TargetNamespaceFilePluginName = "target-namespace-file"
)

type TargetNamespaceFileConfig struct {
	Namespace   string `yaml:"namespace" mapstructure:"namespace"`
	File        string `yaml:"file"  mapstructure:"file"`
	ContentType string `yaml:"content_type"  mapstructure:"content_type"`
}

// TargetNamespaceFilePlugin returns a files in the explorer tree
type TargetNamespaceFilePlugin struct {
	config *TargetNamespaceFileConfig
}

func ConfigureNamespaceFilePlugin(config interface{}, ns string) (plugins.PluginInstance, error) {
	targetNamespaceFileConfig := &TargetNamespaceFileConfig{}

	err := plugins.ConvertConfig(TargetNamespaceFilePluginName, config, targetNamespaceFileConfig)
	if err != nil {
		return nil, err
	}

	// set default to gateway namespace
	if targetNamespaceFileConfig.Namespace == "" {
		targetNamespaceFileConfig.Namespace = ns
	}

	// throw error if non magic namespace targets different namespace
	if targetNamespaceFileConfig.Namespace != ns && ns != core.MagicalGatewayNamespace {
		return nil, fmt.Errorf("plugin can not target different namespace")
	}

	return &TargetNamespaceFilePlugin{
		config: targetNamespaceFileConfig,
	}, nil
}

func (tnf TargetNamespaceFilePlugin) Config() interface{} {
	return tnf.config
}

func (tnf TargetNamespaceFilePlugin) ExecutePlugin(ctx context.Context,
	c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {

	data, err := fetchObjectData(tnf.config.Namespace, tnf.config.File)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
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
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// nolint
	w.Write(data)

	return true
}

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
	Revision struct {
		CreatedAt time.Time `json:"createdAt"`
		Hash      string    `json:"hash"`
		Source    string    `json:"source"`
		Name      string    `json:"name"`
	} `json:"revision"`
	EventLogging string `json:"eventLogging"`
	Oid          string `json:"oid"`
}

func fetchObjectData(ns, path string) ([]byte, error) {

	// prefix with slash if set relative
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	urlString := fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s",
		os.Getenv("DIREKTIV_API_V1_PORT"), ns, path)

	res, err := http.Get(urlString)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var node Node
	err = json.Unmarshal(b, &node)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(node.Revision.Source)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		TargetNamespaceFilePluginName,
		plugins.TargetPluginType,
		ConfigureNamespaceFilePlugin))
}
