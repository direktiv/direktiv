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

	node, data, err := fetchObjectData(resp)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not fetch file data", err)

		return false
	}

	mt := "application/unknown"
	if tnf.config.ContentType != "" {
		mt = tnf.config.ContentType
	} else if node.Data.MimeType != "" {
		mt = node.Data.MimeType
	} else {
		// guessing
		// nolint
		kind, _ := filetype.Match(data)
		if kind != filetype.Unknown {
			mt = kind.MIME.Value
		}
	}
	w.Header().Set("Content-Type", mt)

	// w.Header().Set("Content-Type", mtype.String())
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// nolint
	w.Write(data)

	return true
}

type node struct {
	Data struct {
		Path      string    `json:"path"`
		Type      string    `json:"type"`
		Data      string    `json:"data"`
		Size      int       `json:"size"`
		MimeType  string    `json:"mimeType"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		Children  any       `json:"children"`
	} `json:"data"`
}

func fetchObjectData(res *http.Response) (*node, []byte, error) {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	var node node
	err = json.Unmarshal(b, &node)
	if err != nil {
		return nil, nil, err
	}

	data, err := base64.StdEncoding.DecodeString(node.Data.Data)
	if err != nil {
		return nil, nil, err
	}

	return &node, data, nil
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		NamespaceFilePluginName,
		plugins.TargetPluginType,
		ConfigureNamespaceFilePlugin))
}
