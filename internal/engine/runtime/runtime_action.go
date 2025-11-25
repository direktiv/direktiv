package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/go-viper/mapstructure/v2"
	"github.com/grafana/sobek"
	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func (rt *Runtime) service(t, path string, payload any, retries int) sobek.Value {
	var sd *core.ServiceFileData
	switch t {
	case core.FlowActionScopeSystem:
		sd = &core.ServiceFileData{
			Typ:       core.FlowActionScopeNamespace,
			Namespace: core.FlowActionScopeSystem,
			FilePath:  path,
		}
	case core.FlowActionScopeNamespace:
		sd = &core.ServiceFileData{
			Typ:       core.FlowActionScopeNamespace,
			Namespace: rt.metadata[core.EngineMappingNamespace],
			FilePath:  path,
		}
	default:
		panic(rt.vm.ToValue(fmt.Errorf("unknown scope for script call")))
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("executing service %s in scope %s", path, t))

	var h = make([]map[string]string, 0)
	data, err := rt.callAction(sd, payload, h, retries)
	if err != nil {
		panic(rt.vm.ToValue(err))
	}

	return rt.vm.ToValue(data)
}

func (rt *Runtime) action(c map[string]any) sobek.Value {
	var config core.ActionConfig
	err := mapstructure.Decode(c, &config)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error marshaling action configuration: %s", err.Error())))
	}

	if config.Retries == 0 {
		config.Retries = 2
	}

	config.Type = core.FlowActionScopeLocal

	sd := &core.ServiceFileData{
		Typ:       core.FlowActionScopeLocal,
		Name:      "",
		Namespace: rt.metadata[core.EngineMappingNamespace],
		FilePath:  rt.metadata[core.EngineMappingPath],
		ServiceFile: core.ServiceFile{
			Image: config.Image,
			Cmd:   config.Cmd,
			Size:  config.Size,
			Envs:  config.Envs,
		},
	}
	sd.Name = sd.GetValueHash()

	actionFunc := func(payload any, files []map[string]string) sobek.Value {
		telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
			fmt.Sprintf("executing action with image %s", config.Image))

		data, err := rt.callAction(sd, payload, files, config.Retries)
		if err != nil {
			panic(rt.vm.ToValue(err))
		}

		return rt.vm.ToValue(data)
	}

	return rt.vm.ToValue(actionFunc)
}

func (rt *Runtime) callAction(sd *core.ServiceFileData, payload any, files []map[string]string, retries int) (any, error) {
	rt.onAction(sd.GetID())

	svcUrl := fmt.Sprintf("http://%s.%s.svc", sd.GetID(), os.Getenv("DIREKTIV_SERVICE_NAMESPACE"))

	// ping service
	_, err := callRetryable(rt.tracingPack.ctx, svcUrl+"/up", http.MethodGet, []byte(""), nil, 30)
	if err != nil {
		return nil, fmt.Errorf("action did not start: %s", err.Error())
		// panic(rt.vm.ToValue(fmt.Errorf("action did not start: %s", err.Error())))
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, "action ping successful, calling action")

	data, err := json.Marshal(payload)
	if err != nil {
		// panic(rt.vm.ToValue(fmt.Errorf("could not marshal payload for action: %s", err.Error())))
		return nil, fmt.Errorf("could not marshal payload for action: %s", err.Error())
	}

	h, err := filesToHeader(files)
	if err != nil {
		return nil, fmt.Errorf("could not prepare files for action: %s", err.Error())
	}

	outData, err := callRetryable(rt.tracingPack.ctx, svcUrl, http.MethodPost, data, h, retries)
	if err != nil {
		// panic(rt.vm.ToValue(fmt.Errorf("calling action failed: %s", err.Error())))
		return nil, fmt.Errorf("calling action failed: %s", err.Error())
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, "action call successful")

	var d any
	err = json.Unmarshal(outData, &d)
	if err != nil {
		// panic(rt.vm.ToValue(fmt.Errorf("could not unmarshale response: %s", err.Error())))
		return nil, fmt.Errorf("could not unmarshale response: %s", err.Error())
	}

	return d, nil
}

func filesToHeader(files []map[string]string) (http.Header, error) {
	var h = make(http.Header)
	for a := range files {
		f := files[a]
		scope := "filesystem"
		scope, ok := f["scope"]
		if !ok {
			scope = "filesystem"
		}

		if scope != "filesystem" && scope != "namespace" && scope != "workflow" {
			return h, fmt.Errorf("unknown scope for file for action")
		}

		name, ok := f["name"]
		if !ok {
			return h, fmt.Errorf("name not provided for filr fro action")
		}

		h.Add(core.EngineHeaderFile, fmt.Sprintf("%s;%s", scope, name))
	}

	return h, nil
}

func callRetryable(ctx context.Context, url, method string, payload []byte, headers http.Header, retries int) ([]byte, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = retries
	client.RetryWaitMin = 500 * time.Millisecond
	client.RetryWaitMax = 2 * time.Second
	client.HTTPClient.Timeout = 10 * time.Second // total timeout per request
	// client.Logger = nil                          // silence internal

	req, err := retryablehttp.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	if headers != nil {
		req.Header = headers
	}

	l := ctx.Value(telemetry.DirektivLogCtx(telemetry.LogObjectIdentifier))
	logObject, ok := l.(telemetry.LogObject)
	if !ok {
		return nil, fmt.Errorf("action context missing")
	}

	// set relevant headers
	logObject.ToHeader(&req.Header)

	// inject otel headers for propagation
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO error handling, read headers

	return io.ReadAll(resp.Body)
}
