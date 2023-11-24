package target

// import (
// 	"context"
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/direktiv/direktiv/pkg/refactor/core"
// 	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"

// 	"github.com/h2non/filetype"
// 	"github.com/mitchellh/mapstructure"
// 	"github.com/pkg/errors"
// )

// const (
// 	TargetNamespaceFilePluginName = "target-namespace-file"
// )

// type TargetNamespaceFileConfig struct {
// 	Namespace string `yaml:"namespace"`
// 	File      string `yaml:"file"`
// }

// // TargetNamespaceFilePlugin returns a files in the explorer tree
// type TargetNamespaceFilePlugin struct {
// 	config *TargetNamespaceFileConfig
// }

// func (tnf TargetNamespaceFilePlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
// 	targetNamespaceFileConfig := &TargetNamespaceFileConfig{}

// 	if config != nil {
// 		err := mapstructure.Decode(config, &targetNamespaceFileConfig)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "configuration for target-ns-var invalid")
// 		}
// 	}

// 	// set default to gateway namespace
// 	if targetNamespaceFileConfig.Namespace == "" {
// 		targetNamespaceFileConfig.Namespace = core.MagicalGatewayNamespace
// 	}

// 	return &TargetNamespaceFilePlugin{
// 		config: targetNamespaceFileConfig,
// 	}, nil
// }

// func (tnf TargetNamespaceFilePlugin) Config() interface{} {
// 	return tnf.config
// }

// func (tnf TargetNamespaceFilePlugin) Name() string {
// 	return TargetNamespaceFilePluginName
// }

// func (tnf TargetNamespaceFilePlugin) Type() plugins.PluginType {
// 	return plugins.InboundPluginType
// }

// func (tnf TargetNamespaceFilePlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
// 	w http.ResponseWriter, r *http.Request) bool {

// 	data, err := fetchObjectData(tnf.config.Namespace, tnf.config.File)
// 	if err != nil {
// 		plugins.ReportError(w, http.StatusInternalServerError,
// 			"can not fetch file data", err)
// 		return false
// 	}

// 	r.Header.Set("Content-Type", "application/unknown")

// 	// nolint
// 	kind, _ := filetype.Match(data)
// 	if kind != filetype.Unknown {
// 		r.Header.Set("Content-Type", kind.MIME.Value)
// 	}

// 	// w.Header().Set("Content-Type", mtype.String())
// 	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))

// 	// nolint
// 	w.Write(data)

// 	return true
// }

// //nolint:gochecknoinits
// func init() {
// 	plugins.AddPluginToRegistry(TargetNamespaceFilePlugin{})
// }

// type Node struct {
// 	Namespace string `json:"namespace"`
// 	Node      struct {
// 		CreatedAt    time.Time `json:"createdAt"`
// 		UpdatedAt    time.Time `json:"updatedAt"`
// 		Name         string    `json:"name"`
// 		Path         string    `json:"path"`
// 		Parent       string    `json:"parent"`
// 		Type         string    `json:"type"`
// 		Attributes   []any     `json:"attributes"`
// 		Oid          string    `json:"oid"`
// 		ReadOnly     bool      `json:"readOnly"`
// 		ExpandedType string    `json:"expandedType"`
// 		MimeType     string    `json:"mimeType"`
// 	} `json:"node"`
// 	Revision struct {
// 		CreatedAt time.Time `json:"createdAt"`
// 		Hash      string    `json:"hash"`
// 		Source    string    `json:"source"`
// 		Name      string    `json:"name"`
// 	} `json:"revision"`
// 	EventLogging string `json:"eventLogging"`
// 	Oid          string `json:"oid"`
// }

// func fetchObjectData(ns, path string) ([]byte, error) {

// 	// prefix with slash if set relative
// 	if !strings.HasPrefix(path, "/") {
// 		path = "/" + path
// 	}

// 	urlString := fmt.Sprintf("http://localhost:%s/api/namespaces/%s/tree%s",
// 		os.Getenv("DIREKTIV_API_V1_PORT"), ns, path)

// 	res, err := http.Get(urlString)
// 	if err != nil {
// 		return nil, err
// 	}

// 	b, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	var node Node
// 	err = json.Unmarshal(b, &node)
// 	if err != nil {
// 		return nil, err
// 	}

// 	data, err := base64.StdEncoding.DecodeString(node.Revision.Source)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return data, nil
// }
