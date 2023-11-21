package inbound

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

const (
	RequestConvertPluginName = "request-convert"
)

// RequestConvertConfig converts the whole request into JSON.
type RequestConvertConfig struct {
	OmitHeaders bool `yaml:"omit_headers"`
	OmitQueries bool `yaml:"omit_queries"`
	OmitBody    bool `yaml:"omit_body"`
}

// RequestConvertPlugin converts headers, query parameters, url paramneters
// and the body into a JSON object. The original body is discarded.
type RequestConvertPlugin struct {
	config *RequestConvertConfig
}

func (rcp RequestConvertPlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
	var ok bool
	requestConvertConfig := &RequestConvertConfig{}

	if config != nil {
		requestConvertConfig, ok = config.(*RequestConvertConfig)
		if !ok {
			return nil, fmt.Errorf("configuration for request-convert invalid")
		}
	}

	return &RequestConvertPlugin{
		config: requestConvertConfig,
	}, nil
}

func (rcp RequestConvertPlugin) Name() string {
	return RequestConvertPluginName
}

func (rcp RequestConvertPlugin) Type() plugins.PluginType {
	return plugins.InboundPluginType
}

type RequestConvertResponse struct {
	URLParams   map[string]string   `json:"url-params"`
	QueryParams map[string][]string `json:"query-params"`
	Headers     http.Header         `json:"headers"`
	Body        json.RawMessage     `json:"body"`
}

func (rcp RequestConvertPlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
	w http.ResponseWriter, r *http.Request) bool {

	response := &RequestConvertResponse{
		URLParams:   make(map[string]string),
		QueryParams: make(map[string][]string),
	}

	// convert uri extension
	up := ctx.Value(plugins.URLParamCtxKey)
	if up != nil {
		response.URLParams = up.(map[string]string)
	}

	// convert query params
	if !rcp.config.OmitQueries {
		values := r.URL.Query()
		for k, v := range values {
			response.QueryParams[k] = v
		}
	}

	// convert headers
	if !rcp.config.OmitHeaders {
		response.Headers = r.Header
	}

	// convert content
	var (
		content = []byte("{}")
		err     error
	)
	if r.Body != nil && !rcp.config.OmitBody {
		content, err = io.ReadAll(r.Body)
		if err != nil {
			slog.Error("can not process content",
				slog.String("plugin", RequestConvertPluginName))
			w.WriteHeader(http.StatusBadRequest)
			// nolint
			w.Write([]byte("can not read content"))
			return false
		}
		r.Body.Close()
	}

	// add json content or base64 if binary
	if isJSON(string(content)) {
		response.Body = content
	} else {
		response.Body = []byte(fmt.Sprintf("{ \"data\": \"%s\" }",
			base64.StdEncoding.EncodeToString(content)))
	}

	newBody, err := json.Marshal(response)
	if err != nil {
		slog.Error("can not process content",
			slog.String("plugin", RequestConvertPluginName))
		w.WriteHeader(http.StatusInternalServerError)
		// nolint
		w.Write([]byte("can not marshal content"))
		return false
	}
	r.Body = io.NopCloser(bytes.NewBuffer(newBody))

	slog.Debug("converted content set",
		slog.String("plugin", RequestConvertPluginName),
		slog.String("body", string(newBody)))

	return true
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(RequestConvertPlugin{})
}
