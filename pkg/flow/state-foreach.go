package flow

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/vorteil/direktiv/pkg/model"
)

type foreachStateLogic struct {
	state    *model.ForEachState
	workflow *model.Workflow
}

func initForEachStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	foreach, ok := state.(*model.ForEachState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(foreachStateLogic)
	sl.state = foreach
	sl.workflow = wf

	return sl, nil

}

func (sl *foreachStateLogic) Type() string {
	return model.StateTypeForEach.String()
}

func (sl *foreachStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return deadlineFromString(ctx, engine, im, sl.state.Timeout)
}

func (sl *foreachStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *foreachStateLogic) ID() string {
	return sl.state.ID
}

func (sl *foreachStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {

	var err error
	var children = make([]stateChild, 0)

	var logics []multiactionTuple
	err = im.UnmarshalMemory(&logics)
	if err != nil {
		engine.sugar.Error(err)
		return children
	}

	for _, logic := range logics {
		if logic.Complete {
			continue
		}
		children = append(children, stateChild{
			Id:   logic.ID,
			Type: logic.Type,
		})
	}

	return children

}

func (sl *foreachStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *foreachStateLogic) do(ctx context.Context, engine *engine, im *instanceMemory, inputSource interface{}, attempt int) (logic multiactionTuple, err error) {

	action := sl.state.Action

	var inputData []byte
	inputData, err = generateActionInput(ctx, engine, im, inputSource, action)
	if err != nil {
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

		sf := fn.(*model.SubflowFunctionDefinition)

		// subflow

		caller := new(subflowCaller)
		caller.InstanceID = im.ID().String()
		caller.State = sl.state.GetID()
		caller.Step = im.Step()

		var subflowID string

		// TODO: log subflow instance IDs

		subflowID, err = engine.subflowInvoke(ctx, caller, im.in.Edges.Namespace, sf.Workflow, inputData)
		if err != nil {
			return
		}

		logic = multiactionTuple{
			ID:       subflowID,
			Type:     "subflow",
			Attempts: attempt,
		}
	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.GlobalKnativeFunctionType:
		fallthrough
	case model.ReusableContainerFunctionType:

		uid := ksuid.New()
		logic = multiactionTuple{
			ID:       uid.String(),
			Type:     "isolate",
			Attempts: attempt,
		}

		var ar *functionRequest
		ar, err = engine.newIsolateRequest(ctx, im, sl.state.GetID(), 0, fn, inputData, uid, false)
		if err != nil {
			return
		}

		err = engine.doActionRequest(ctx, ar)
		if err != nil {
			return
		}

	case model.IsolatedContainerFunctionType:

		uid := ksuid.New()
		logic = multiactionTuple{
			ID:       uid.String(),
			Type:     "isolate",
			Attempts: attempt,
		}

		var ar *functionRequest
		ar, err = engine.newIsolateRequest(ctx, im, sl.state.GetID(), 0, fn, inputData, uid, false)
		if err != nil {
			return
		}

		err = engine.doActionRequest(ctx, ar)
		if err != nil {
			return
		}
	default:
		err = NewInternalError(fmt.Errorf("unsupported function type: %v", fnt))
		return
	}

	return

}

func (sl *foreachStateLogic) doAll(ctx context.Context, engine *engine, im *instanceMemory) (err error) {

	var array []interface{}
	array, err = jq(im.data, sl.state.Array)
	if err != nil {
		return
	}

	engine.logToInstance(ctx, time.Now(), im.in, "Generated %d objects to loop over.", len(array))

	logics := make([]multiactionTuple, 0)

	for _, inputSource := range array {
		var logic multiactionTuple
		logic, err = sl.do(ctx, engine, im, inputSource, 0)
		if err != nil {
			return
		}
		logics = append(logics, logic)
	}

	err = engine.SetMemory(ctx, im, logics)
	if err != nil {
		return
	}

	return

}

func (sl *foreachStateLogic) doSpecific(ctx context.Context, engine *engine, im *instanceMemory, logics []multiactionTuple, idx int) (err error) {

	var array []interface{}
	array, err = jq(im.data, sl.state.Array)
	if err != nil {
		return
	}

	inputSource := array[idx]

	var logic multiactionTuple
	logic, err = sl.do(ctx, engine, im, inputSource, logics[idx].Attempts)
	if err != nil {
		return
	}
	logics[idx] = logic

	err = engine.SetMemory(ctx, im, logics)
	if err != nil {
		return
	}

	return

}

func (sl *foreachStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if im.GetMemory() != nil {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		err = sl.doAll(ctx, engine, im)
		if err != nil {
			return
		}

		return

	}

	var logics []multiactionTuple
	err = im.UnmarshalMemory(&logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	// check for scheduled retry
	retryData := new(foreachStateLogicRetry)
	dec := json.NewDecoder(bytes.NewReader(wakedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(retryData)
	if err == nil {
		engine.logToInstance(ctx, time.Now(), im.in, "Retrying...")
		err = sl.doSpecific(ctx, engine, im, logics, retryData.Idx)
		return
	}

	// second part
	results := new(actionResultPayload)
	err = json.Unmarshal(wakedata, results)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	var found bool
	var idx int
	var completed int

	for i, lid := range logics {

		if lid.Complete {
			completed++
		}

		if lid.ID == results.ActionID {
			found = true
			if lid.Complete {
				err = NewInternalError(fmt.Errorf("action '%s' already completed", lid.ID))
				return
			}
			idx = i
		}

	}

	if !found {
		err = NewInternalError(fmt.Errorf("action '%s' wasn't expected", results.ActionID))
		return
	}

	if results.ErrorCode != "" {

		err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
		engine.logToInstance(ctx, time.Now(), im.in, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
		var d time.Duration
		d, err = preprocessRetry(sl.state.Action.Retries, logics[idx].Attempts, err)
		if err != nil {
			return
		}

		engine.logToInstance(ctx, time.Now(), im.in, "Scheduling retry attempt in: %v.", d)
		err = sl.scheduleRetry(ctx, engine, im, logics, idx, d)
		return

	}

	if results.ErrorMessage != "" {
		engine.logToInstance(ctx, time.Now(), im.in, "Action crashed due to an internal error: %v", results.ErrorMessage)
		err = NewInternalError(errors.New(results.ErrorMessage))
		return
	}

	logics[idx].Complete = true
	completed++
	engine.logToInstance(ctx, time.Now(), im.in, "Action returned. (%d/%d)", completed, len(logics))

	var x interface{}
	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	logics[idx].Results = x

	var ready bool
	if completed == len(logics) {
		ready = true
	}

	if ready {

		var results []interface{}
		for i := range logics {
			results = append(results, logics[i].Results)
		}

		err = im.StoreData("return", results)
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

	err = engine.SetMemory(ctx, im, logics)
	if err != nil {
		return
	}

	return

}

type foreachStateLogicRetry struct {
	Logics []multiactionTuple
	Idx    int
}

func (r *foreachStateLogicRetry) Marshal() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return data
}

func (sl *foreachStateLogic) scheduleRetry(ctx context.Context, engine *engine, im *instanceMemory, logics []multiactionTuple, idx int, d time.Duration) error {

	var err error

	logics[idx].Attempts++
	logics[idx].ID = ""

	err = engine.SetMemory(ctx, im, logics)
	if err != nil {
		return err
	}

	r := &foreachStateLogicRetry{
		Idx:    idx,
		Logics: logics,
	}
	data := r.Marshal()

	t := time.Now().Add(d)

	err = engine.scheduleRetry(im.ID().String(), sl.ID(), im.Step(), t, data)
	if err != nil {
		return err
	}

	return nil

}
