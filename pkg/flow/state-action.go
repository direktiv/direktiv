package flow

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/senseyeio/duration"
	"github.com/vorteil/direktiv/pkg/functions"
	igrpc "github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/model"
)

type actionStateLogic struct {
	state    *model.ActionState
	workflow *model.Workflow
}

func initActionStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	action, ok := state.(*model.ActionState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(actionStateLogic)
	sl.state = action
	sl.workflow = wf

	return sl, nil

}

func (sl *actionStateLogic) Type() string {
	return model.StateTypeAction.String()
}

func (sl *actionStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {

	if sl.state.Async {
		return time.Now().Add(time.Second * 5)
	}

	var t time.Time
	var d time.Duration

	d = time.Minute * 15

	if sl.state.Timeout != "" {
		dur, err := duration.ParseISO8601(sl.state.Timeout)
		if err != nil {
			engine.logToInstance(ctx, time.Now(), im.in, "Got an invalid ISO8601 timeout: %v", err)
		} else {
			now := time.Now()
			later := dur.Shift(now)
			d = later.Sub(now)
		}
	}

	t = time.Now()
	t = t.Add(d)
	t = t.Add(time.Second * 5)

	return t

}

func (sl *actionStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *actionStateLogic) ID() string {
	return sl.state.ID
}

func (sl *actionStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {

	var err error
	var children = make([]stateChild, 0)

	sd := new(actionStateSavedata)
	err = im.UnmarshalMemory(sd)
	if err != nil {
		engine.sugar.Error(err)
		return children
	}

	if sl.state.Action.Function != "" && sd.Id != "" {

		var uid ksuid.KSUID
		err = uid.UnmarshalText([]byte(sd.Id))
		if err != nil {
			engine.sugar.Error(err)
			return children
		}

		children = append(children, stateChild{
			Id:   uid.String(),
			Type: "isolate",
		})

	} else {

		id := string(sd.Id)

		children = append(children, stateChild{
			Id:   id,
			Type: "subflow",
		})

	}

	return children

}

func (sl *actionStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

type actionStateSavedata struct {
	Op       string
	Id       string
	Attempts int
}

func (sd *actionStateSavedata) Marshal() []byte {
	data, err := json.Marshal(sd)
	if err != nil {
		panic(err)
	}
	return data
}

func (engine *engine) newIsolateRequest(ctx context.Context, im *instanceMemory, stateId string, timeout int,
	fn model.FunctionDefinition, inputData []byte,
	uid ksuid.KSUID, async bool) (*functionRequest, error) {

	wf, err := engine.InstanceWorkflow(ctx, im)
	if err != nil {
		return nil, err
	}

	ar := new(functionRequest)
	ar.ActionID = uid.String()
	// ar.Workflow.Name = wli.wf.Name
	ar.Workflow.WorkflowID = wf.ID.String()
	ar.Workflow.Timeout = timeout

	ar.Workflow.NamespaceName = im.in.Edges.Namespace.Name
	ar.Workflow.Path = im.in.As

	if !async {
		ar.Workflow.InstanceID = im.ID().String()
		ar.Workflow.NamespaceID = im.in.Edges.Namespace.ID.String()
		ar.Workflow.State = stateId
		ar.Workflow.Step = im.Step()
	}

	// TODO: timeout
	fnt := fn.GetType()
	ar.Container.Type = fnt
	ar.Container.Data = inputData

	wfID := im.in.Edges.Workflow.ID.String()
	revID := im.in.Edges.Revision.ID.String()
	nsID := im.in.Edges.Namespace.ID.String()

	switch fnt {
	case model.ReusableContainerFunctionType:
		con := fn.(*model.ReusableFunctionDefinition)
		ar.Container.Image = con.Image
		ar.Container.Cmd = con.Cmd
		ar.Container.Size = con.Size
		ar.Container.Scale = con.Scale
		ar.Container.ID = con.ID
		ar.Container.Service, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name:      &con.ID,
			Workflow:  &wfID,
			Revision:  &revID,
			Namespace: &nsID,
		})
		if err != nil {
			panic(err)
		}
		ar.Container.Files = con.Files
	case model.IsolatedContainerFunctionType:
		con := fn.(*model.IsolatedFunctionDefinition)
		ar.Container.Image = con.Image
		ar.Container.Cmd = con.Cmd
		ar.Container.Size = con.Size
		ar.Container.ID = con.ID
		ar.Container.Files = con.Files
	case model.NamespacedKnativeFunctionType:
		con := fn.(*model.NamespacedFunctionDefinition)
		ar.Container.Files = con.Files
		ar.Container.ID = con.ID
		ar.Container.Service, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name:      &con.ID,
			Namespace: &nsID,
		})
	case model.GlobalKnativeFunctionType:
		con := fn.(*model.GlobalFunctionDefinition)
		ar.Container.Files = con.Files
		ar.Container.ID = con.ID
		ar.Container.Service, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name: &con.ID,
		})
	default:
		return nil, fmt.Errorf("unexpected function type: %v", fn)
	}

	return ar, nil

}

func (sl *actionStateLogic) do(ctx context.Context, engine *engine, im *instanceMemory, attempt int) (transition *stateTransition, err error) {

	var inputData []byte
	inputData, err = generateActionInput(ctx, engine, im, im.data, sl.state.Action)
	if err != nil {
		return
	}

	// default 15 mins timeout
	wfto := 15 * 60
	if len(sl.state.Timeout) > 0 {
		var to duration.Duration
		to, err = duration.ParseISO8601(sl.state.Timeout)
		if err != nil {
			return
		}
		dur := to.Shift(time.Now()).Sub(time.Now())
		wfto = int(dur.Seconds())
	}

	fn, err := sl.workflow.GetFunction(sl.state.Action.Function)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	fnt := fn.GetType()
	switch fnt {
	case model.SubflowFunctionType:

		sf := fn.(*model.SubflowFunctionDefinition)

		caller := new(subflowCaller)
		caller.InstanceID = im.ID().String()
		caller.State = sl.state.GetID()
		caller.Step = im.Step()

		var subflowID string

		subflowID, err = engine.subflowInvoke(ctx, caller, im.in.Edges.Namespace, sf.Workflow, inputData)
		if err != nil {
			return
		}

		if sl.state.Async {
			engine.logToInstance(ctx, time.Now(), im.in, "Running subflow '%s' in fire-and-forget mode (async).", subflowID)
			transition = &stateTransition{
				Transform: sl.state.Transform,
				NextState: sl.state.Transition,
			}
			return
		} else {
			engine.logToInstance(ctx, time.Now(), im.in, "Sleeping until subflow '%s' returns.", subflowID)
			sd := &actionStateSavedata{
				Op:       "do",
				Id:       subflowID,
				Attempts: attempt,
			}
			err = engine.SetMemory(ctx, im, sd)
			if err != nil {
				return
			}
		}
	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.GlobalKnativeFunctionType:
		fallthrough
	case model.ReusableContainerFunctionType:

		uid := ksuid.New()

		sd := &actionStateSavedata{
			Op:       "do",
			Id:       uid.String(),
			Attempts: attempt,
		}

		err = engine.SetMemory(ctx, im, sd)
		if err != nil {
			return
		}

		var ar *functionRequest
		ar, err = engine.newIsolateRequest(ctx, im, sl.state.GetID(), wfto, fn, inputData, uid, sl.state.Async)
		if err != nil {
			return
		}

		if sl.state.Async {
			engine.logToInstance(ctx, time.Now(), im.in, "Running function '%s' in fire-and-forget mode (async).", fn.GetID())
			go func(ctx context.Context, im *instanceMemory, ar *functionRequest) {
				err = engine.doActionRequest(ctx, ar)
				if err != nil {
					return
				}
			}(ctx, im, ar)
			transition = &stateTransition{
				Transform: sl.state.Transform,
				NextState: sl.state.Transition,
			}
			return
		} else {
			engine.logToInstance(ctx, time.Now(), im.in, "Sleeping until function '%s' returns.", fn.GetID())
			err = engine.doActionRequest(ctx, ar)
			if err != nil {
				return
			}
		}

	case model.IsolatedContainerFunctionType:

		uid := ksuid.New()

		sd := &actionStateSavedata{
			Op:       "do",
			Id:       uid.String(),
			Attempts: attempt,
		}

		err = engine.SetMemory(ctx, im, sd)
		if err != nil {
			return
		}

		var ar *functionRequest
		ar, err = engine.newIsolateRequest(ctx, im, sl.state.GetID(), wfto, fn, inputData, uid, sl.state.Async)
		if err != nil {
			return
		}

		if sl.state.Async {
			engine.logToInstance(ctx, time.Now(), im.in, "Running function '%s' in fire-and-forget mode (async).", fn.GetID())
			go func(ctx context.Context, im *instanceMemory, ar *functionRequest) {
				err = engine.doActionRequest(ctx, ar)
				if err != nil {
					return
				}
			}(ctx, im, ar)
			transition = &stateTransition{
				Transform: sl.state.Transform,
				NextState: sl.state.Transition,
			}
			return
		} else {
			engine.logToInstance(ctx, time.Now(), im.in, "Sleeping until function '%s' returns.", fn.GetID())
			err = engine.doActionRequest(ctx, ar)
			if err != nil {
				return
			}
		}
	default:
		err = NewInternalError(fmt.Errorf("unsupported function type: %v", fnt))
		return
	}

	return
}

func (sl *actionStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if im.GetMemory() != nil {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		return sl.do(ctx, engine, im, 0)

	}

	// check for scheduled retry
	retryData := new(actionStateSavedata)
	dec := json.NewDecoder(bytes.NewReader(wakedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(retryData)
	if err == nil && retryData.Op == "retry" {
		engine.logToInstance(ctx, time.Now(), im.in, "Retrying...")
		return sl.do(ctx, engine, im, retryData.Attempts)
	}

	// second part

	results := new(actionResultPayload)
	dec = json.NewDecoder(bytes.NewReader(wakedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(results)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	sd := new(actionStateSavedata)
	err = im.UnmarshalMemory(sd)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	fn, err := sl.workflow.GetFunction(sl.state.Action.Function)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	fnt := fn.GetType()
	switch fnt {
	case model.SubflowFunctionType:
		id := sd.Id
		if results.ActionID != id {
			err = NewInternalError(errors.New("incorrect subflow action ID"))
			return
		}
		engine.logToInstance(ctx, time.Now(), im.in, "Subflow '%s' returned.", id)
	case model.ReusableContainerFunctionType:
		fallthrough
	case model.IsolatedContainerFunctionType:
		fallthrough
	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.GlobalKnativeFunctionType:
		var uid ksuid.KSUID
		err = uid.UnmarshalText([]byte(sd.Id))
		if err != nil {
			err = NewInternalError(err)
			return
		}

		if results.ActionID != uid.String() {
			err = NewInternalError(errors.New("incorrect action ID"))
			return
		}

		engine.logToInstance(ctx, time.Now(), im.in, "Function '%s' returned.", sl.state.Action.Function)
	default:
		err = NewInternalError(fmt.Errorf("unexpected function type: %v", fn))
		return
	}

	if results.ErrorCode != "" {

		err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
		engine.logToInstance(ctx, time.Now(), im.in, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
		var d time.Duration

		d, err = preprocessRetry(sl.state.Action.Retries, sd.Attempts, err)
		if err != nil {
			return
		}

		engine.logToInstance(ctx, time.Now(), im.in, "Scheduling retry attempt in: %v.", d)
		err = sl.scheduleRetry(ctx, engine, im, sd, d)
		return

	}

	if results.ErrorMessage != "" {

		engine.logToInstance(ctx, time.Now(), im.in, "Action crashed due to an internal error: %v", results.ErrorMessage)

		err = NewInternalError(errors.New(results.ErrorMessage))
		return
	}

	var x interface{}
	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	err = im.StoreData("return", x)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}

func (sl *actionStateLogic) scheduleRetry(ctx context.Context, engine *engine, im *instanceMemory, sd *actionStateSavedata, d time.Duration) error {

	var err error

	sd.Attempts++
	sd.Op = "retry"
	sd.Id = ""

	err = engine.SetMemory(ctx, im, sd)
	if err != nil {
		return err
	}

	data := sd.Marshal()

	t := time.Now().Add(d)

	err = engine.scheduleRetry(im.ID().String(), sl.ID(), im.Step(), t, data)
	if err != nil {
		return err
	}

	return nil

}

func generateActionInput(ctx context.Context, engine *engine, im *instanceMemory, data interface{}, action *model.ActionDefinition) ([]byte, error) {

	var err error
	var input interface{}

	input, err = jqObject(data, "jq(.)")
	if err != nil {
		return nil, err
	}

	m, ok := input.(map[string]interface{})
	if !ok {
		err = NewInternalError(errors.New("invalid state data"))
		return nil, err
	}

	m, err = addSecrets(ctx, engine, im, m, action.Secrets...)
	if err != nil {
		return nil, err
	}

	if action.Input == nil {
		input, err = jqOne(m, "jq(.)")
		if err != nil {
			return nil, err
		}
	} else {
		input, err = jqOne(m, action.Input)
		if err != nil {
			return nil, err
		}
	}

	var inputData []byte

	inputData, err = json.Marshal(input)
	if err != nil {
		err = NewInternalError(err)
		return nil, err
	}

	return inputData, nil

}

func isRetryable(code string, patterns []string) bool {

	for _, pattern := range patterns {
		// NOTE: this error should be checked in model validation

		if pattern == "*" {
			pattern = ".*"
		}

		matched, _ := regexp.MatchString(pattern, code)
		if matched {
			return true
		}
	}

	return false

}

func retryDelay(attempt int, delay string, multiplier float64) time.Duration {

	d := time.Second * 5
	if x, err := duration.ParseISO8601(delay); err == nil {
		t0 := time.Now()
		t1 := x.Shift(t0)
		d = t1.Sub(t0)
	}

	if multiplier != 0 {
		for i := 0; i < attempt; i++ {
			d = time.Duration(float64(d) * multiplier)
		}
	}

	return d

}

func preprocessRetry(retry *model.RetryDefinition, attempt int, err error) (time.Duration, error) {

	var d time.Duration

	if retry == nil {
		return d, err
	}

	cerr, ok := err.(*CatchableError)
	if !ok {
		return d, err
	}

	if !isRetryable(cerr.Code, retry.Codes) {
		return d, err
	}

	if attempt >= retry.MaxAttempts {
		return d, NewCatchableError("direktiv.retries.exceeded", "maximum retries exceeded")
	}

	d = retryDelay(attempt, retry.Delay, retry.Multiplier)

	return d, nil

}
