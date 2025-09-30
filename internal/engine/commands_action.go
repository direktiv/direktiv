package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/service"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/grafana/sobek"
	"github.com/hashicorp/go-retryablehttp"
)

func doubleMarshal[T any](in any) (*T, error) {
	data, ok := in.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("action image configuration has wrong type")
	}

	j, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("action image configuration can not be converted: %s", err.Error())
	}

	var outData T
	if err := json.Unmarshal(j, &outData); err != nil {
		return nil, fmt.Errorf("action image configuration can not be converted: %s", err.Error())
	}

	return &outData, err
}

func (cmds *Commands) action(call sobek.FunctionCall) sobek.Value {
	if len(call.Arguments) != 1 {
		panic(cmds.vm.ToValue("action definition needs configuration"))
	}

	actionConfig, err := doubleMarshal[core.ActionConfig](call.Argument(0).ToObject(cmds.vm).Export())
	if err != nil {
		panic(err)
	}

	actionFunc := func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) == 0 {
			panic(cmds.vm.ToValue("data element is missing in action payload"))
		}
		var retValue any

		payload, err := doubleMarshal[core.ActionPayload](call.Argument(0).ToObject(cmds.vm).Export())
		if err != nil {
			panic(cmds.vm.ToValue(err))
		}

		// TODO: ad namespace services
		switch actionConfig.Type {
		case core.FlowActionScopeLocal:
			err = cmds.callLocal(actionConfig, payload)
			if err != nil {
				panic(cmds.vm.ToValue(err))
			}
		case core.FlowActionScopeSubflow:
		default:
			panic(cmds.vm.ToValue(fmt.Sprintf("unknown action type '%s'", actionConfig.Type)))
		}

		return cmds.vm.ToValue(retValue)
	}

	return cmds.vm.ToValue(actionFunc)
}

const (
	DirektivActionIDHeader     = "Direktiv-ActionID"
	DirektivTempDir            = "Direktiv-TempDir"
	DirektivErrorCodeHeader    = "Direktiv-ErrorCode"
	DirektivErrorMessageHeader = "Direktiv-ErrorMessage"
)

func (cmds *Commands) callLocal(config *core.ActionConfig, payload *core.ActionPayload) error {
	// TODO: remove this
	{
		err := cmds.deletemeStartServiceManager()
		if err != nil {
			return err
		}
		err = cmds.deleteme(config, payload)
		if err != nil {
			return err
		}
	}

	// convert the data
	dataBinary, err := json.Marshal(payload.Data)
	if err != nil {
		return err
	}

	// generate url
	svcURL := cmdsSm.GetServiceURL(cmds.metadata[core.EngineMappingNamespace],
		core.FlowActionScopeLocal, cmds.metadata[core.EngineMappingPath], "dummy")

	slog.Info("requesting action", slog.String("url", svcURL))

	// post the data
	req, err := retryablehttp.NewRequest(http.MethodPost, svcURL,
		bytes.NewReader(dataBinary))
	if err != nil {
		return err
	}

	// set action id header
	req.Header.Set(core.EngineHeaderActionID, cmds.instID.String())

	for i := range payload.Files {
		req.Header.Add(core.EngineHeaderFile, payload.Files[i])
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	retryClient.RequestLogHook = func(l retryablehttp.Logger, r *http.Request, i int) {
		slog.Info("retrying action call", slog.String("url", svcURL))
	}
	retryClient.RetryMax = 50

	resp, err := retryClient.Do(req)
	if err != nil {
		return err
	}

	b, _ := httputil.DumpResponse(resp, true)
	fmt.Println(string(b))

	return nil
}

// TODO: delete me
var cmdsSm core.ServiceManager

func (cmds *Commands) deletemeStartServiceManager() error {
	if cmdsSm != nil {
		return nil
	}

	fas := func() ([]string, error) {
		return []string{}, nil
	}

	config := &core.Config{}
	if err := env.Parse(config); err != nil {
		return fmt.Errorf("parsing env variables: %w", err)
	}
	if err := config.Init(); err != nil {
		return fmt.Errorf("init config, err: %w", err)
	}

	config.FunctionsReconcileInterval = 1
	sm, err := service.NewManager(config, fas)
	if err != nil {
		return err
	}

	go sm.Start(lifecycle.New(context.TODO(), os.Interrupt))

	cmdsSm = sm

	return nil
}

func (cmds *Commands) deleteme(config *core.ActionConfig, payload any) error {
	svd := &core.ServiceFileData{
		Typ:       core.FlowActionScopeLocal,
		Namespace: cmds.metadata[core.EngineMappingNamespace],
		FilePath:  cmds.metadata[core.EngineMappingPath],
		Name:      "dummy",
	}

	if config.Inject {
		svd.Cmd = "/usr/share/direktiv/direktiv-cmd"
	}

	envVars := make([]core.EnvironmentVariable, 0)
	for k, v := range config.Envs {
		envVars = append(envVars, core.EnvironmentVariable{
			Name:  k,
			Value: v,
		})
	}

	svd.Size = config.Size
	svd.Image = config.Image

	cmdsSm.SetServices([]*core.ServiceFileData{
		svd,
	})

	// GetServiceURL(namespace string, typ string, file string, name string) string
	surl := cmdsSm.GetServiceURL(cmds.metadata[core.EngineMappingNamespace], core.FlowActionScopeLocal, cmds.metadata[core.EngineMappingPath], "dummy")

	for range make([]int, 20) {
		err := cmdsSm.IgniteService(surl)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}
