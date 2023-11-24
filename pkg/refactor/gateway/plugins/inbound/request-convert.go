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

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	RequestConvertPluginName = "request-convert"
)

// RequestConvertConfig converts the whole request into JSON.
type RequestConvertConfig struct {
	OmitHeaders  bool `yaml:"omit_headers" mapstructure:"omit_headers"`
	OmitQueries  bool `yaml:"omit_queries" mapstructure:"omit_queries"`
	OmitBody     bool `yaml:"omit_body" mapstructure:"omit_body"`
	OmitConsumer bool `yaml:"omit_consumer" mapstructure:"omit_consumer"`
}

// RequestConvertPlugin converts headers, query parameters, url paramneters
// and the body into a JSON object. The original body is discarded.
type RequestConvertPlugin struct {
	config *RequestConvertConfig
}

func ConfigureRequestConvert(config interface{}, ns string) (plugins.PluginInstance, error) {
	requestConvertConfig := &RequestConvertConfig{}

	if config != nil {
		err := mapstructure.Decode(config, &requestConvertConfig)
		if err != nil {
			return nil, errors.Wrap(err, "configuration for request-convert invalid")
		}
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
	URLParams   map[string]string   `json:"url-params"`
	QueryParams map[string][]string `json:"query-params"`
	Headers     http.Header         `json:"headers"`
	Body        json.RawMessage     `json:"body"`
	Consumer    RequestConsumer     `json:"consumer"`
}

func (rcp *RequestConvertPlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {

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

func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		RequestConvertPluginName,
		plugins.InboundPluginType,
		ConfigureRequestConvert))
}
