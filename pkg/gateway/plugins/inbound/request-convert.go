package inbound

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

// RequestConvertPlugin converts headers, query parameters, url paramneters
// and the body into a JSON object. The original body is discarded.
type RequestConvertPlugin struct {
	OmitHeaders  bool `mapstructure:"omit_headers"`
	OmitQueries  bool `mapstructure:"omit_queries"`
	OmitBody     bool `mapstructure:"omit_body"`
	OmitConsumer bool `mapstructure:"omit_consumer"`
	OmitMethod   bool `mapstructure:"omit_method"`
}

func (rcp *RequestConvertPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &RequestConvertPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

type RequestConsumer struct {
	Username string   `json:"username"`
	Tags     []string `json:"tags"`
	Groups   []string `json:"groups"`
}

type RequestConvertResponse struct {
	URLParams   map[string]string   `json:"url_params"`
	QueryParams map[string][]string `json:"query_params"`
	Headers     http.Header         `json:"headers"`
	Body        json.RawMessage     `json:"body"`
	Consumer    RequestConsumer     `json:"consumer"`
	Method      string              `json:"method"`
}

func (rcp *RequestConvertPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	response := &RequestConvertResponse{
		URLParams:   make(map[string]string),
		QueryParams: make(map[string][]string),
		Consumer: RequestConsumer{
			Username: "",
			Tags:     make([]string, 0),
			Groups:   make([]string, 0),
		},
	}

	// add url param
	response.URLParams = make(map[string]string)
	for _, param := range gateway.ExtractContextURLParams(r) {
		response.URLParams[param] = r.PathValue(param)
	}

	// convert query params
	if !rcp.OmitQueries {
		values := r.URL.Query()
		for k, v := range values {
			response.QueryParams[k] = v
		}
	}

	// convert headers
	if !rcp.OmitHeaders {
		response.Headers = r.Header
	}

	// convert method
	if !rcp.OmitMethod {
		response.Method = r.Method
	}
	c := gateway.ExtractContextActiveConsumer(r)

	if !rcp.OmitConsumer && c != nil {
		response.Consumer.Username = c.Username
		response.Consumer.Tags = c.Tags
		response.Consumer.Groups = c.Groups
	}

	// convert content
	var (
		content = []byte("{}")
		err     error
	)
	if r.Body != nil && !rcp.OmitBody {
		content, err = io.ReadAll(r.Body)
		if err != nil {
			gateway.WriteInternalError(r, w, err, "can not process content")
			return nil, nil
		}
		defer r.Body.Close()
	}

	// add json content or base64 if binary
	if gateway.IsJSON(string(content)) {
		response.Body = content
	} else {
		response.Body = []byte(fmt.Sprintf("{ \"data\": \"%s\" }",
			base64.StdEncoding.EncodeToString(content)))
	}

	newBody, err := json.Marshal(response)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not process content")
		return nil, nil
	}
	r.Body = io.NopCloser(bytes.NewBuffer(newBody))

	slog.Debug("converted content set",
		"plugin", (&RequestConvertPlugin{}).Type(),
		"body", string(newBody))

	return w, r
}

func (rcp *RequestConvertPlugin) Type() string {
	return "request-convert"
}

func init() {
	gateway.RegisterPlugin(&RequestConvertPlugin{})
}
