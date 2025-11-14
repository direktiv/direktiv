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

func (rt *Runtime) action(c map[string]any) sobek.Value {
	var config core.ActionConfig
	err := mapstructure.Decode(c, &config)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error marshaling action configuration: %s", err.Error())))
	}

	if config.Retries == 0 {
		config.Retries = 2
	}

	sd := &core.ServiceFileData{
		Typ:       core.ServiceTypeWorkflow,
		Name:      "",
		Namespace: rt.metadata[core.EngineMappingNamespace],
		FilePath:  rt.metadata[core.EngineMappingPath],
		ServiceFile: core.ServiceFile{
			Image: config.Image,
			Cmd:   config.Cmd,
			Size:  config.Size,
			Envs:  config.Envs,
			// TODO: this need to be set to zero to enable zero scaling.
			Scale: 1,
		},
	}
	sd.Name = sd.GetValueHash()

	actionFunc := func(payload any) sobek.Value {
		telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo,
			fmt.Sprintf("executing action with image %s", config.Image))

		rt.onAction(sd.GetID())

		svcUrl := fmt.Sprintf("http://%s.%s.svc", sd.GetID(), os.Getenv("DIREKTIV_SERVICE_NAMESPACE"))

		// ping service
		_, err := callRetryable(rt.tracingPack.ctx, svcUrl+"/up", http.MethodGet, []byte(""), 30)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("action did not start: %s", err.Error())))
		}

		telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, "action ping successful, calling action")

		data, err := json.Marshal(payload)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("could not marshal payload for action: %s", err.Error())))
		}

		outData, err := callRetryable(rt.tracingPack.ctx, svcUrl, http.MethodPost, data, config.Retries)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("calling action failed: %s", err.Error())))
		}

		telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, "action call successful")

		var d any
		err = json.Unmarshal(outData, &d)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("could not unmarshale response: %s", err.Error())))
		}

		return rt.vm.ToValue(d)
	}

	return rt.vm.ToValue(actionFunc)
}

func callRetryable(ctx context.Context, url, method string, payload []byte, retries int) ([]byte, error) {
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
