package outbound

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
	"github.com/dop251/goja"
)

type JSOutboundPlugin struct {
	Script string `mapstructure:"script" yaml:"script"`
}

func (js *JSOutboundPlugin) NewInstance(_ core.EndpointV2, config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &JSOutboundPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

type response struct {
	Headers http.Header
	Body    string
	Code    int
}

func (js *JSOutboundPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	var (
		err error
		b   []byte
	)

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("can not set read body for js plugin")
		}
		defer r.Body.Close()
	}

	if r.Response == nil {
		r.Response = &http.Response{
			StatusCode: http.StatusOK,
		}
	}

	resp := response{
		Headers: r.Header,
		Body:    string(b),
		Code:    r.Response.StatusCode,
	}

	// extract all response headers and body
	vm := goja.New()
	err = vm.Set("input", resp)
	if err != nil {
		return nil, fmt.Errorf("can not set input object")
	}

	err = vm.Set("log", func(txt interface{}) {
		slog.Info("js log", slog.Any("log", txt))
	})
	if err != nil {
		return nil, fmt.Errorf("can not set log function")
	}

	err = vm.Set("sleep", func(t interface{}) {
		tt, ok := t.(int64)
		if !ok {
			return
		}
		time.Sleep(time.Duration(tt) * time.Second)
	})
	if err != nil {
		return nil, fmt.Errorf("can not set sleep function")
	}

	script := fmt.Sprintf("function run() { %s; return input } run()",
		js.Script)

	val, err := vm.RunScript("plugin", script)
	if err != nil {
		return nil, fmt.Errorf("can not execute script")
	}

	if val != nil && !val.Equals(goja.Undefined()) {
		o := val.ToObject(vm)
		// make sure the input object got returned
		if o.ExportType() == reflect.TypeOf(resp) {
			// nolint checked before
			responseDone := o.Export().(response)
			for k, v := range responseDone.Headers {
				for a := range v {
					w.Header().Add(k, v[a])
				}
			}

			// nolint
			w.Write([]byte(responseDone.Body))
			w.WriteHeader(responseDone.Code)
		}
	}

	return r, nil
}

func (js *JSOutboundPlugin) Type() string {
	return "js-outbound"
}

func init() {
	plugins.RegisterPlugin(&JSOutboundPlugin{})
}
