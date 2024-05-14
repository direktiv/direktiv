package inbound

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
	"github.com/dop251/goja"
)

// JSInboundPlugin allows to modify headers, contents and query params of the request.
type JSInboundPlugin struct {
	Script string `mapstructure:"script" yaml:"script"`
}

func (js *JSInboundPlugin) NewInstance(_ core.EndpointV2, config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &JSInboundPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (js *JSInboundPlugin) Type() string {
	return "js-inbound"
}

type Query struct {
	U url.Values
}

func (q Query) Get(key string) []string {
	return q.U[key]
}

func (q Query) Set(key string, value string) {
	q.U.Set(key, value)
}

func (q Query) Add(key string, value string) {
	q.U.Add(key, value)
}

func (q Query) Delete(key string) {
	q.U.Del(key)
}

type Values struct {
	Something map[string][]string
}

type request struct {
	Headers http.Header
	// Headers shared.Headers
	Queries url.Values
	// Queries shared.Query
	Body string

	Consumer *core.ConsumerFileV2

	// url params of type /{id}
	URLParams map[string]string

	// response code after executing
	Status int
}

func (js *JSInboundPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	var (
		err error
		b   []byte
	)

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("can not set read body for js inbound plugin")
		}
		defer r.Body.Close()
	}

	vm := goja.New()

	var c *core.ConsumerFileV2
	if gateway2.ParseRequestActiveConsumer(r) != nil {
		c = &gateway2.ParseRequestActiveConsumer(r).ConsumerFileV2
	} else {
		c = &core.ConsumerFileV2{}
	}

	// add url param
	urlParams := make(map[string]string)

	// TODO: fix here.
	// up := r.Context().Value(plugins.URLParamCtxKey)
	// if up != nil {
	//	// nolint we know it is from us
	//	urlParams = up.(map[string]string)
	//}

	req := request{
		Headers:   r.Header,
		Queries:   r.URL.Query(),
		Body:      string(b),
		Consumer:  c,
		URLParams: urlParams,
		Status:    0,
	}

	// extract all response headers and body
	err = vm.Set("input", req)
	if err != nil {
		return nil, fmt.Errorf("can not set input object")
	}

	err = vm.Set("log", func(txt interface{}) {
		slog.Info("js log", slog.Any("log", txt))
	})
	if err != nil {
		return nil, fmt.Errorf("can not set log function")
	}

	script := fmt.Sprintf("function run() { %s; return input } run()",
		js.Script)

	val, err := vm.RunScript("plugin", script)
	if err != nil {
		return nil, fmt.Errorf("can not execute script")
	}

	if val != nil && !val.Equals(goja.Undefined()) {
		r.Header = http.Header{}

		o := val.ToObject(vm)
		// make sure the input object got returned
		if o.ExportType() == reflect.TypeOf(req) {
			// nolint checked before
			responseDone := o.Export().(request)
			addHeader(responseDone.Headers, r.Header)

			newQuery := make(url.Values)
			for k, v := range responseDone.Queries {
				for a := range v {
					newQuery.Add(k, v[a])
				}
			}
			r.URL.RawQuery = newQuery.Encode()
			r.Body = io.NopCloser(strings.NewReader(responseDone.Body))

			// TODO: discuss with jens
			// script set status code and stop executing other plugins
			// if responseDone.Status > 0 {
			//	return serveResponse(w, responseDone)
			//}
		}
	}

	return r, nil
}

func init() {
	plugins.RegisterPlugin(&JSInboundPlugin{})
}

func addHeader(getHeader, setHeader http.Header) {
	for k, v := range getHeader {
		for a := range v {
			setHeader.Add(k, v[a])
		}
	}
}
