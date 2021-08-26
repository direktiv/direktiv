package direktiv

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

func (sl *actionStateLogic) Deadline() time.Time {

	if sl.state.Async {
		return time.Now().Add(time.Second * 5)
	}

	var t time.Time
	var d time.Duration

	d = time.Minute * 15

	if sl.state.Timeout != "" {
		dur, err := duration.ParseISO8601(sl.state.Timeout)
		if err != nil {
			// NOTE: validation should prevent this from ever happening
			appLog.Errorf("Got an invalid ISO8601 timeout: %v", err)
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

func (sl *actionStateLogic) LivingChildren(savedata []byte) []stateChild {

	var err error
	var children = make([]stateChild, 0)

	sd := new(actionStateSavedata)
	err = json.Unmarshal(savedata, sd)
	if err != nil {
		appLog.Error(err)
		return children
	}

	if sl.state.Action.Function != "" {

		var uid ksuid.KSUID
		err = uid.UnmarshalText([]byte(sd.Id))
		if err != nil {
			appLog.Error(err)
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

func (wli *workflowLogicInstance) newIsolateRequest(stateId string, timeout int,
	fn model.FunctionDefinition, inputData []byte,
	uid ksuid.KSUID, async bool) (*functionRequest, error) {

	ar := new(functionRequest)
	ar.ActionID = uid.String()
	ar.Workflow.Name = wli.wf.Name
	ar.Workflow.ID = wli.wf.ID
	ar.Workflow.Timeout = timeout

	if !async {
		ar.Workflow.InstanceID = wli.id
		ar.Workflow.Namespace = wli.namespace
		ar.Workflow.State = stateId
		ar.Workflow.Step = wli.step
	}

	// TODO: timeout
	fnt := fn.GetType()
	ar.Container.Type = fnt
	ar.Container.Data = inputData

	switch fnt {
	case model.ReusableContainerFunctionType:
		con := fn.(*model.ReusableFunctionDefinition)
		ar.Container.Image = con.Image
		ar.Container.Cmd = con.Cmd
		ar.Container.Size = con.Size
		ar.Container.Scale = con.Scale
		ar.Container.ID = con.ID
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
		ar.Container.Service = fmt.Sprintf("%s-%s-%s", functions.PrefixNamespace, wli.namespace, con.KnativeService)
	case model.GlobalKnativeFunctionType:
		con := fn.(*model.GlobalFunctionDefinition)
		ar.Container.Files = con.Files
		ar.Container.ID = con.ID
		ar.Container.Service = fmt.Sprintf("%s-%s", functions.PrefixGlobal, con.KnativeService)
	default:
		return nil, fmt.Errorf("unexpected function type: %v", fn)
	}

	return ar, nil

}

func (sl *actionStateLogic) do(ctx context.Context, instance *workflowLogicInstance, attempt int) (transition *stateTransition, err error) {

	var inputData []byte
	inputData, err = generateActionInput(ctx, instance, instance.data, sl.state.Action)
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
		caller.InstanceID = instance.id
		caller.State = sl.state.GetID()
		caller.Step = instance.step

		var subflowID string

		subflowID, err = instance.engine.subflowInvoke(ctx, caller, instance.rec.InvokedBy, instance.namespace, sf.Workflow, inputData)
		if err != nil {
			return
		}

		if sl.state.Async {
			instance.Log(ctx, "Running subflow '%s' in fire-and-forget mode (async).", subflowID)
			transition = &stateTransition{
				Transform: sl.state.Transform,
				NextState: sl.state.Transition,
			}
			return
		} else {
			instance.Log(ctx, "Sleeping until subflow '%s' returns.", subflowID)
			sd := &actionStateSavedata{
				Op:       "do",
				Id:       subflowID,
				Attempts: attempt,
			}
			err = instance.Save(ctx, sd.Marshal())
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

		err = instance.Save(ctx, sd.Marshal())
		if err != nil {
			return
		}

		var ar *functionRequest
		ar, err = instance.newIsolateRequest(sl.state.GetID(), wfto, fn, inputData, uid, sl.state.Async)
		if err != nil {
			return
		}

		if sl.state.Async {
			instance.Log(ctx, "Running function '%s' in fire-and-forget mode (async).", fn.GetID())
			go func(ctx context.Context, instance *workflowLogicInstance, ar *functionRequest) {
				err = instance.engine.doActionRequest(ctx, ar)
				if err != nil {
					return
				}
			}(ctx, instance, ar)
			transition = &stateTransition{
				Transform: sl.state.Transform,
				NextState: sl.state.Transition,
			}
			return
		} else {
			instance.Log(ctx, "Sleeping until function '%s' returns.", fn.GetID())
			err = instance.engine.doActionRequest(ctx, ar)
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

		err = instance.Save(ctx, sd.Marshal())
		if err != nil {
			return
		}

		var ar *functionRequest
		ar, err = instance.newIsolateRequest(sl.state.GetID(), wfto, fn, inputData, uid, sl.state.Async)
		if err != nil {
			return
		}

		if sl.state.Async {
			instance.Log(ctx, "Running function '%s' in fire-and-forget mode (async).", fn.GetID())
			go func(ctx context.Context, instance *workflowLogicInstance, ar *functionRequest) {
				err = instance.engine.doActionRequest(ctx, ar)
				if err != nil {
					return
				}
			}(ctx, instance, ar)
			transition = &stateTransition{
				Transform: sl.state.Transform,
				NextState: sl.state.Transition,
			}
			return
		} else {
			instance.Log(ctx, "Sleeping until function '%s' returns.", fn.GetID())
			err = instance.engine.doActionRequest(ctx, ar)
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

func (sl *actionStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		return sl.do(ctx, instance, 0)

	}

	// check for scheduled retry
	retryData := new(actionStateSavedata)
	dec := json.NewDecoder(bytes.NewReader(wakedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(retryData)
	if err == nil && retryData.Op == "retry" {
		instance.Log(ctx, "Retrying...")
		return sl.do(ctx, instance, retryData.Attempts)
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
	dec = json.NewDecoder(bytes.NewReader(savedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(sd)
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
		instance.Log(ctx, "Subflow '%s' returned.", id)
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

		instance.Log(ctx, "Function '%s' returned.", sl.state.Action.Function)
	default:
		err = NewInternalError(fmt.Errorf("unexpected function type: %v", fn))
		return
	}

	if results.ErrorCode != "" {

		err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
		instance.Log(ctx, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
		var d time.Duration

		d, err = preprocessRetry(sl.state.Action.Retries, sd.Attempts, err)
		if err != nil {
			return
		}

		instance.Log(ctx, "Scheduling retry attempt in: %v.", d)
		err = sl.scheduleRetry(ctx, instance, sd, d)
		return

	}

	if results.ErrorMessage != "" {

		instance.Log(ctx, "Action crashed due to an internal error: %v", results.ErrorMessage)

		err = NewInternalError(errors.New(results.ErrorMessage))
		return
	}

	var x interface{}
	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	err = instance.StoreData("return", x)
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

func (sl *actionStateLogic) scheduleRetry(ctx context.Context, instance *workflowLogicInstance, sd *actionStateSavedata, d time.Duration) error {

	var err error

	sd.Attempts++
	sd.Op = "retry"
	sd.Id = ""

	data := sd.Marshal()
	err = instance.Save(ctx, data)
	if err != nil {
		return err
	}

	t := time.Now().Add(d)

	err = instance.engine.scheduleRetry(instance.id, sl.ID(), instance.step, t, data)
	if err != nil {
		return err
	}

	return nil

}

func generateActionInput(ctx context.Context, instance *workflowLogicInstance, data interface{}, action *model.ActionDefinition) ([]byte, error) {

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

	m, err = addSecrets(ctx, instance, m, action.Secrets...)
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
