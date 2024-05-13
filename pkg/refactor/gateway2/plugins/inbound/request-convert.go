package inbound

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const (
	RequestConvertPluginName = "request-convert"
)

// RequestConvertConfig converts the whole request into JSON.
type RequestConvertConfig struct {
	OmitHeaders  bool `mapstructure:"omit_headers"  yaml:"omit_headers"`
	OmitQueries  bool `mapstructure:"omit_queries"  yaml:"omit_queries"`
	OmitBody     bool `mapstructure:"omit_body"     yaml:"omit_body"`
	OmitConsumer bool `mapstructure:"omit_consumer" yaml:"omit_consumer"`
}

// RequestConvertPlugin converts headers, query parameters, url paramneters
// and the body into a JSON object. The original body is discarded.
type RequestConvertPlugin struct {
	config *RequestConvertConfig
}

func (rcp *RequestConvertPlugin) Construct(config core.PluginConfigV2) (core.PluginV2, error) {
	requestConvertConfig := &RequestConvertConfig{}

	err := plugins.ConvertConfig(config.Config, requestConvertConfig)
	if err != nil {
		return nil, err
	}

	return &RequestConvertPlugin{
		config: requestConvertConfig,
	}, nil
}

func (rcp *RequestConvertPlugin) Config() interface{} {
	return rcp.config
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
}

func (rcp *RequestConvertPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	response := &RequestConvertResponse{
		URLParams:   make(map[string]string),
		QueryParams: make(map[string][]string),
		Consumer: RequestConsumer{
			Username: "",
			Tags:     make([]string, 0),
			Groups:   make([]string, 0),
		},
	}

	// convert uri extension
	up := r.Context().Value(plugins.URLParamCtxKey)
	if up != nil {
		// nolint cvoming from gateway
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

	if !rcp.config.OmitConsumer && c != nil {
		response.Consumer.Username = c.Username
		response.Consumer.Tags = c.Tags
		response.Consumer.Groups = c.Groups
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
	if plugins.IsJSON(string(content)) {
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

func (rcp *RequestConvertPlugin) Type() string {
	return RequestConvertPluginName
}

func init() {
	plugins.RegisterPlugin(&RequestConvertPlugin{})
}
