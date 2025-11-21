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
		},
	}
	sd.Name = sd.GetValueHash()

	actionFunc := func(actionCallArgs map[string]any) sobek.Value {
		telemetry.LogInstance(rt.ctx, telemetry.LogLevelInfo,
			fmt.Sprintf("executing action with image %s", config.Image))

		rt.onAction(sd.GetID())

		if _, ok := actionCallArgs["body"]; !ok {
			panic(rt.vm.ToValue(fmt.Errorf("action call args missing 'body' field ")))
		}
		headers := make(map[string]string)
		if _, ok := actionCallArgs["headers"]; ok {
			err := fmt.Errorf("action call args 'headers' field should be an object mapping string keys to string values")
			h, ok := actionCallArgs["headers"].(map[string]any)
			if !ok {
				panic(rt.vm.ToValue(err))
			}
			for k, v := range h {
				switch s := v.(type) {
				case string:
					headers[k] = s
				default:
					panic(rt.vm.ToValue(err))
				}
			}
		}

		svcUrl := fmt.Sprintf("http://%s.%s.svc", sd.GetID(), os.Getenv("DIREKTIV_SERVICE_NAMESPACE"))

		// ping service
		_, err := callRetryable(rt.ctx, svcUrl+"/up", http.MethodGet, nil, []byte(""), 30)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("action did not start: %s", err.Error())))
		}

		telemetry.LogInstance(rt.ctx, telemetry.LogLevelInfo, "action ping successful, calling action")

		data, err := json.Marshal(actionCallArgs["body"])
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("could not marshal payload for action: %s", err.Error())))
		}
		outData, err := callRetryable(rt.ctx, svcUrl, http.MethodPost, headers, data, config.Retries)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("calling action failed: %s", err.Error())))
		}

		telemetry.LogInstance(rt.ctx, telemetry.LogLevelInfo, "action call successful")

		var d any
		err = json.Unmarshal(outData, &d)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Errorf("could not unmarshale response: %s", err.Error())))
		}

		return rt.vm.ToValue(d)
	}

	return rt.vm.ToValue(actionFunc)
}

func callRetryable(ctx context.Context, url, method string, headers map[string]string, payload []byte, retries int) ([]byte, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = retries
	client.RetryWaitMin = 500 * time.Millisecond
	client.RetryWaitMax = 2 * time.Second
	client.HTTPClient.Timeout = 10 * time.Second // total timeout per request
	// client.Logger = nil                          // silence internal

	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO error handling, read headers

	return io.ReadAll(resp.Body)
}
