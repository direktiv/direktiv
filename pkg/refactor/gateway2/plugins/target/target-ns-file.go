package target

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
)

// NamespaceFilePlugin returns a files in the explorer tree.
type NamespaceFilePlugin struct {
	Namespace string `mapstructure:"namespace"`
	File      string `mapstructure:"file"`
}

func (tnf *NamespaceFilePlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &NamespaceFilePlugin{}

	err := gateway2.ConvertConfig(config.Config, pl)
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
	if resp.StatusCode != http.StatusOK {
		gateway2.WriteInternalError(r, w, nil, "couldn't execute downstream request")
		return nil
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't read downstream request")
		return nil
	}

	type PayLoad struct {
		Data struct {
			Data     string `json:"data"`
			Typ      string `json:"type"`
			MimeType string `json:"mimeType"`
		} `json:"data"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}

	payLoad := &PayLoad{}
	err = json.Unmarshal(b, payLoad)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't decode downstream response")
		return nil
	}
	if payLoad.Error.Code != "" {
		gateway2.WriteInternalError(r, w, nil, "downstream response error")
		return nil
	}
	if payLoad.Data.Typ == string(filestore.FileTypeDirectory) {
		gateway2.WriteInternalError(r, w, nil, "requested file is a directory")
		return nil
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(payLoad.Data.Data)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't base64 decode downstream response")
		return nil
	}
	w.Header().Set("Content-Type", payLoad.Data.MimeType)

	_, err = w.Write(decodedBytes)
	if err != nil {
		gateway2.WriteInternalError(r, w, nil, "couldn't write downstream response")
		return nil
	}

	return r
}

func init() {
	gateway2.RegisterPlugin(&NamespaceFilePlugin{})
}
