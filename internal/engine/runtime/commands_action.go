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

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/service"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/hashicorp/go-retryablehttp"
)

type actionError struct {
	code    string
	message string
}

func (ae *actionError) Error() string {
	return "THIS IS ACTION ERROR"
}

func (ae *actionError) PanicObject(vm *sobek.Runtime) *sobek.Object {
	errorConstructor := vm.Get("Error")
	errorInstance, _ := vm.New(errorConstructor, vm.ToValue(ae.message))

	errorObj := errorInstance.ToObject(vm)
	errorObj.Set("code", ae.code)
	errorObj.Set("message", ae.message)

	return errorObj
}

func newActionLogger(ctx context.Context) *actionLogger {
	// get values from context
	// if reqID, ok := ctx.Value("request_id").(string); ok {
	// 	logger = logger.With("request_id", reqID)
	// }
	// if userID, ok := ctx.Value("user_id").(string); ok {
	// 	logger = logger.With("user_id", userID)
	// }
	// if traceID, ok := ctx.Value("trace_id").(string); ok {
	// 	logger = logger.With("trace_id", traceID)
	// }

	return &actionLogger{}
}

func (n *actionLogger) Error(msg string, keysAndValues ...interface{}) {
	fmt.Println(msg)
}

func (n *actionLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Println(msg)
}

func (n *actionLogger) Debug(msg string, keysAndValues ...interface{}) {
	fmt.Println(msg)
}

func (n *actionLogger) Warn(msg string, keysAndValues ...interface{}) {
	fmt.Println(msg)
}

func (cmds *Runtime) action(call sobek.FunctionCall) sobek.Value {
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

		// TODO: add namespace services, system services
		switch actionConfig.Type {
		case core.FlowActionScopeLocal:
			data, err := cmds.callLocal(actionConfig, payload)
			if err != nil {
				actionError := &actionError{}
				if errors.As(err, &actionError) {
					panic(actionError.PanicObject(cmds.vm))
				} else {
					panic(cmds.vm.ToValue(err))
				}
			}

			return cmds.vm.ToValue(data)
		case core.FlowActionScopeSubflow:
		default:
			panic(cmds.vm.ToValue(fmt.Sprintf("unknown action type '%s'", actionConfig.Type)))
		}

		return cmds.vm.ToValue(retValue)
	}

	// returns the action to be called in the script later
	return cmds.vm.ToValue(actionFunc)
}

type ActionCaller struct {
	instID  uuid.UUID
	addr    string
	payload *core.ActionPayload

	errMessage, errCode string
}

func (ac *ActionCaller) call(ctx context.Context) (any, error) {
	// do up check, /up is the test path in the action service
	req, err := retryablehttp.NewRequest(http.MethodPost, fmt.Sprintf("%s/up", ac.addr), nil)
	if err != nil {
		return nil, err
	}

	_, err = ac.doRequest(ctx, req, 60, 1)
	if err != nil {
		return nil, err
	}

	slog.Info("service is up", slog.String("svc", ac.addr))

	// convert the data
	dataBinary, err := json.Marshal(ac.payload.Data)
	if err != nil {
		return nil, err
	}

	req, err = retryablehttp.NewRequest(http.MethodPost, ac.addr, bytes.NewReader(dataBinary))
	if err != nil {
		return nil, err
	}

	// set action header
	req.Header.Set(core.EngineHeaderActionID, ac.instID.String())

	// set files header
	for i := range ac.payload.Files {
		req.Header.Add(core.EngineHeaderFile, ac.payload.Files[i])
	}

	if ac.payload.RetryInterval == 0 {
		ac.payload.RetryInterval = 1
	}

	data, err := ac.doRequest(ctx, req, ac.payload.Retries, ac.payload.RetryInterval)
	if err != nil {
		if ac.errMessage != "" {
			actionErr := &actionError{
				code:    ac.errCode,
				message: ac.errMessage,
			}

			return nil, actionErr
		}

		return nil, err
	}

	var retObject any
	err = json.Unmarshal(data, &retObject)
	if err != nil {
		return nil, err
	}

	return retObject, nil
}

type actionLogger struct {
	instID uuid.UUID
	logger *slog.Logger
}

func (ac *ActionCaller) doRequest(ctx context.Context, req *retryablehttp.Request, retry, interval int) ([]byte, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retry
	retryClient.RetryWaitMin = time.Duration(interval) * time.Second
	retryClient.RequestLogHook = func(l retryablehttp.Logger, r *http.Request, i int) {
		telemetry.LogInstance(ctx, telemetry.LogLevelInfo, fmt.Sprintf("executing request to %s", ac.addr))
	}

	retryClient.Logger = newActionLogger(ctx)

	retryClient.ResponseLogHook = func(l retryablehttp.Logger, r *http.Response) {
		// set the error if there is one in the header
		ac.errCode = r.Header.Get(core.EngineHeaderErrorCode)
		ac.errMessage = r.Header.Get(core.EngineHeaderErrorMessage)
	}

	resp, err := retryClient.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return data, err
}

func (cmds *Runtime) callLocal(config *core.ActionConfig, payload *core.ActionPayload) (any, error) {
	// TODO: remove this
	{
		err := cmds.deletemeStartServiceManager()
		if err != nil {
			return "", err
		}
		err = cmds.deleteme(config, payload)
		if err != nil {
			return "", err
		}
	}

	// TODO: name needs to be hashed image, cmd, envs etc.
	svcURL := cmdsSm.GetServiceURL(cmds.metadata[core.EngineMappingNamespace],
		core.FlowActionScopeLocal, cmds.metadata[core.EngineMappingPath], "dummy")

	slog.Debug("requesting action", slog.String("url", svcURL))

	actionCaller := &ActionCaller{
		instID:  cmds.instID,
		addr:    svcURL,
		payload: payload,
	}

	return actionCaller.call(context.Background())
}

// TODO: delete me.
var cmdsSm core.ServiceManager

func (cmds *Runtime) deletemeStartServiceManager() error {
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

func (cmds *Runtime) deleteme(config *core.ActionConfig, payload any) error {
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
