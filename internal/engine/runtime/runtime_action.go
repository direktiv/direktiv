package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/go-viper/mapstructure/v2"
	"github.com/grafana/sobek"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sosodev/duration"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const defaultTimeout = 5 * time.Minute

type fileSpec struct {
	Name       string
	Scope      string
	Permission string
}

func (rt *Runtime) parseFiles(raw any) []fileSpec {
	filesArr, ok := raw.([]any)
	if !ok {
		panic(rt.vm.ToValue(fmt.Errorf("files must be an array")))
	}

	var specs []fileSpec

	for _, f := range filesArr {
		fileMap, ok := f.(map[string]any)
		if !ok {
			panic(rt.vm.ToValue(fmt.Errorf("each file entry must be an object with name and scope")))
		}

		name, ok := fileMap["name"].(string)
		if !ok || name == "" {
			panic(rt.vm.ToValue(fmt.Errorf("file entry must have a non-empty 'name' string")))
		}

		scope, ok := fileMap["scope"].(string)
		if !ok || scope == "" {
			panic(rt.vm.ToValue(fmt.Errorf("file entry must have a non-empty 'scope' string")))
		}

		permission, _ := fileMap["permission"].(string)

		switch scope {
		case "namespace", "workflow", "file":
			telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
				fmt.Sprintf("adding file '%s' in scope '%s'", name, scope))

			specs = append(specs, fileSpec{Name: name, Scope: scope, Permission: permission})
		default:
			telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
				fmt.Sprintf("unknown file scope '%s' for variable '%s', skipping", scope, name))
		}
	}

	return specs
}

func (rt *Runtime) service(c map[string]any) sobek.Value {
	// func (rt *Runtime) service(t, path string, payload any, retries int) sobek.Value {

	t, ok := c["scope"]
	if !ok {
		panic(rt.vm.ToValue(fmt.Errorf("scope not provided, must be namespace or system")))
	}

	p, ok := c["path"]
	if !ok {
		panic(rt.vm.ToValue(fmt.Errorf("path not provided, must be set to service file in namespace or system")))
	}

	path, ok := p.(string)
	if !ok {
		panic(rt.vm.ToValue(fmt.Errorf("path must be a string")))
	}

	r, ok := c["retries"]
	if !ok {
		r = any(int64(0))
	}

	retries, ok := r.(int64)
	if !ok {
		panic(rt.vm.ToValue(fmt.Errorf("retries must be an integer")))
	}

	endDuration := defaultTimeout

	to, ok := c["timeout"]
	if ok {
		timeout, ok := to.(string)
		if !ok {
			panic(rt.vm.ToValue(fmt.Errorf("timeout must be a string")))
		}

		to, err := duration.Parse(timeout)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("timeout not a valid ISO8601 string, e.g. PT1M")))
		}

		endDuration = to.ToTimeDuration()
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("service call timeout in %s", endDuration.String()))

	payload, ok := c["payload"]
	if !ok {
		payload = ""
	}

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

	// parse file specs to pass along as headers to the sidecar
	var files []fileSpec
	if rawFiles, ok := c["files"]; ok {
		files = rt.parseFiles(rawFiles)
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("executing service %s in scope %s", path, t))

	data, err := rt.callAction(sd, payload, int(retries), endDuration, nil, files)
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

	config.Type = core.FlowActionScopeWorkflow

	sd := &core.ServiceFileData{
		Typ:       core.FlowActionScopeWorkflow,
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

	actionTimeout := config.Timeout

	actionFunc := func(call sobek.FunctionCall) sobek.Value {
		var payload any
		if len(call.Arguments) > 0 {
			if err := rt.vm.ExportTo(call.Arguments[0], &payload); err != nil {
				panic(rt.vm.ToValue(fmt.Sprintf("error exporting action payload: %s", err.Error())))
			}
		}

		var (
			callTimeout string
			files       []fileSpec
		)

		// second argument is an optional options object: { timeout: "PT15M", files: [...] }
		if len(call.Arguments) > 1 {
			var opts map[string]any
			if err := rt.vm.ExportTo(call.Arguments[1], &opts); err != nil {
				panic(rt.vm.ToValue(fmt.Sprintf("error exporting action options: %s", err.Error())))
			}

			if t, ok := opts["timeout"].(string); ok {
				callTimeout = t
			}
			if rawFiles, ok := opts["files"]; ok && rawFiles != nil {
				files = rt.parseFiles(rawFiles)
			}
		}

		telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
			fmt.Sprintf("executing action with image %s", config.Image))

		endDuration := defaultTimeout

		effectiveTimeout := actionTimeout
		if callTimeout != "" {
			effectiveTimeout = callTimeout
		}

		if effectiveTimeout != "" {
			to, err := duration.Parse(effectiveTimeout)
			if err != nil {
				panic(rt.vm.ToValue(fmt.Errorf("timeout not a valid ISO8601 string, e.g. PT1M")))
			}

			endDuration = to.ToTimeDuration()
		}

		telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
			fmt.Sprintf("action timeout in %s", endDuration.String()))

		data, err := rt.callAction(sd, payload, config.Retries, endDuration, config.Auth, files)
		if err != nil {
			panic(rt.vm.ToValue(err))
		}

		return rt.vm.ToValue(data)
	}

	return rt.vm.ToValue(actionFunc)
}

func filesToHeaders(headers *http.Header, files []fileSpec) {
	for _, f := range files {
		v := fmt.Sprintf("%s:%s", f.Scope, f.Name)
		if f.Permission != "" {
			v = fmt.Sprintf("%s:%s:%s", f.Scope, f.Name, f.Permission)
		}
		headers.Add(core.EngineHeaderFile, v)
	}
}

func (rt *Runtime) callAction(sd *core.ServiceFileData, payload any, retries int, timeout time.Duration, auth *core.BasicAuthConfig, files []fileSpec) (any, error) {
	if rt.onAction != nil {
		err := rt.onAction(sd.GetID())
		if err != nil {
			return nil, err
		}
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, "connecting to action/service")

	svcUrl := fmt.Sprintf("http://%s.%s.svc", sd.GetID(), os.Getenv("DIREKTIV_SERVICE_NAMESPACE"))

	// ping service
	_, err := callRetryable(rt.tracingPack.ctx, svcUrl+"/up", http.MethodGet, []byte(""), 29, 10*time.Second, auth, nil)
	if err != nil {
		slog.Error("could not connect to service or action", slog.Any("error", err))
		return nil, fmt.Errorf("cannot connect to service or action, please check action/service deployment")
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, "action ping successful, calling action")

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not marshal payload for action: %s", err.Error())
	}

	outData, err := callRetryable(rt.tracingPack.ctx, svcUrl, http.MethodPost, data, retries, timeout, auth, files)
	if err != nil {
		slog.Error("could not call service or action", slog.Any("error", err))
		return nil, fmt.Errorf("calling action failed: %s", err.Error())
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, "action call successful")

	var d any
	err = json.Unmarshal(outData, &d)
	if err != nil {
		slog.Error("could not unmarshal response", slog.Any("error", err), slog.String("data", string(data)))
		return nil, fmt.Errorf("could not unmarshal response: %s", err.Error())
	}

	return d, nil
}

func callRetryable(ctx context.Context, url, method string, payload []byte, retries int, timeout time.Duration, auth *core.BasicAuthConfig, files []fileSpec) ([]byte, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = retries
	client.RetryWaitMin = 500 * time.Millisecond
	client.RetryWaitMax = 2 * time.Second
	client.HTTPClient.Timeout = timeout // total timeout per request
	// client.Logger = nil                          // silence internal

	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	l := ctx.Value(telemetry.DirektivLogCtx(telemetry.LogObjectIdentifier))
	logObject, ok := l.(telemetry.LogObject)
	if !ok {
		return nil, fmt.Errorf("action context missing")
	}

	// set relevant headers
	logObject.ToHeader(&req.Header)

	// inject otel headers for propagation
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// set file headers for the sidecar
	filesToHeaders(&req.Header, files)

	if auth != nil {
		req.SetBasicAuth(auth.Username, auth.Password)
	}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("request exceeded timeout of %s", timeout.String())
		}
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("request cancelled: %w", err)
		}

		return nil, err
	}
	defer resp.Body.Close()

	// TODO error handling, read headers

	return io.ReadAll(resp.Body)
}
