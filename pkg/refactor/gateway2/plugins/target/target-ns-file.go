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

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/h2non/filetype"
)

// NamespaceFilePlugin returns a files in the explorer tree.
type NamespaceFilePlugin struct {
	Namespace   string `mapstructure:"namespace"    yaml:"namespace"`
	File        string `mapstructure:"file"         yaml:"file"`
	ContentType string `mapstructure:"content_type" yaml:"content_type"`
}

func (tnf *NamespaceFilePlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &NamespaceFilePlugin{}

	err := plugins.ConvertConfig(config, pl)
	if err != nil {
		return nil, err
	}

	if pl.File == "" {
		return nil, fmt.Errorf("file is required")
	}

	if !strings.HasPrefix(pl.File, "/") {
		pl.File = "/" + pl.File
	}

	return pl, nil
}

func (tnf *NamespaceFilePlugin) Type() string {
	return "target-namespace-file"
}

func (tnf *NamespaceFilePlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	currentNS := gateway2.ExtractContextEndpoint(r).Namespace
	if tnf.Namespace == "" {
		tnf.Namespace = currentNS
	}
	if tnf.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway2.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil
	}

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s",
		os.Getenv("DIREKTIV_API_V2_PORT"), tnf.Namespace, tnf.File)

	// request failed if nil and response already written
	resp, err := doRequest(r, http.MethodGet, url, nil)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	defer resp.Body.Close()

	data, mime, err := fetchObjectData(resp)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "can not fetch file data")
		return nil
	}

	mt := "application/unknown"

	// overwrite object mimetype if configured
	// otherwise use the one coming from the API
	// last resort is guessing
	if tnf.ContentType != "" {
		mt = tnf.ContentType
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

	return r
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

func init() {
	gateway2.RegisterPlugin(&NamespaceFilePlugin{})
}
