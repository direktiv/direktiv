package direktiv

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
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

func (sl *foreachStateLogic) Deadline() time.Time {
	return deadlineFromString(sl.state.Timeout)
}

func (sl *foreachStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *foreachStateLogic) ID() string {
	return sl.state.ID
}

func (sl *foreachStateLogic) LivingChildren(savedata []byte) []stateChild {

	var err error
	var children = make([]stateChild, 0)

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
	if err != nil {
		log.Error(err)
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

func (sl *foreachStateLogic) do(ctx context.Context, instance *workflowLogicInstance, inputSource interface{}, attempt int) (logic multiactionTuple, err error) {

	action := sl.state.Action

	var inputData []byte
	inputData, err = generateActionInput(ctx, instance, inputSource, action)
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
		caller.InstanceID = instance.id
		caller.State = sl.state.GetID()
		caller.Step = instance.step

		var subflowID string

		// TODO: log subflow instance IDs

		subflowID, err = instance.engine.subflowInvoke(ctx, caller, instance.rec.InvokedBy, instance.namespace, sf.Workflow, inputData)
		if err != nil {
			return
		}

		logic = multiactionTuple{
			ID:       subflowID,
			Type:     "subflow",
			Attempts: attempt,
		}

	case model.ReusableContainerFunctionType:

		uid := ksuid.New()
		logic = multiactionTuple{
			ID:       uid.String(),
			Type:     "isolate",
			Attempts: attempt,
		}

		var ar *isolateRequest
		ar, err = instance.newIsolateRequest(sl.state.GetID(), 0, fn, inputData, uid, false)
		if err != nil {
			return
		}

		err = instance.engine.doActionRequest(ctx, ar)
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

		var ar *isolateRequest
		ar, err = instance.newIsolateRequest(sl.state.GetID(), 0, fn, inputData, uid, false)
		if err != nil {
			return
		}

		err = instance.engine.doActionRequest(ctx, ar)
		if err != nil {
			return
		}

	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.GlobalKnativeFunctionType:
		fallthrough
	default:
		err = NewInternalError(fmt.Errorf("unsupported function type: %v", fnt))
		return
	}

	return

}

func (sl *foreachStateLogic) doAll(ctx context.Context, instance *workflowLogicInstance) (err error) {

	var array []interface{}
	array, err = jq(instance.data, sl.state.Array)
	if err != nil {
		return
	}

	instance.Log(ctx, "Generated %d objects to loop over.", len(array))

	logics := make([]multiactionTuple, 0)

	for _, inputSource := range array {
		var logic multiactionTuple
		logic, err = sl.do(ctx, instance, inputSource, 0)
		if err != nil {
			return
		}
		logics = append(logics, logic)
	}

	var data []byte
	data, err = json.Marshal(logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	err = instance.Save(ctx, data)
	if err != nil {
		return
	}

	return

}

func (sl *foreachStateLogic) doSpecific(ctx context.Context, instance *workflowLogicInstance, logics []multiactionTuple, idx int) (err error) {

	var array []interface{}
	array, err = jq(instance.data, sl.state.Array)
	if err != nil {
		return
	}

	inputSource := array[idx]

	var logic multiactionTuple
	logic, err = sl.do(ctx, instance, inputSource, logics[idx].Attempts)
	if err != nil {
		return
	}
	logics[idx] = logic

	var data []byte
	data, err = json.Marshal(logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	err = instance.Save(ctx, data)
	if err != nil {
		return
	}

	return

}

func (sl *foreachStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		err = sl.doAll(ctx, instance)
		if err != nil {
			return
		}

		return

	}

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
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
		instance.Log(ctx, "Retrying...")
		err = sl.doSpecific(ctx, instance, logics, retryData.Idx)
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
		instance.Log(ctx, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
		var d time.Duration
		d, err = preprocessRetry(sl.state.Action.Retries, logics[idx].Attempts, err)
		if err != nil {
			return
		}

		instance.Log(ctx, "Scheduling retry attempt in: %v.", d)
		err = sl.scheduleRetry(ctx, instance, logics, idx, d)
		return

	}

	if results.ErrorMessage != "" {
		instance.Log(ctx, "Action crashed due to an internal error: %v", results.ErrorMessage)
		err = NewInternalError(errors.New(results.ErrorMessage))
		return
	}

	logics[idx].Complete = true
	completed++
	instance.Log(ctx, "Action returned. (%d/%d)", completed, len(logics))

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

		err = instance.StoreData("return", results)
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

	var data []byte
	data, err = json.Marshal(logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	err = instance.Save(ctx, data)
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

func (sl *foreachStateLogic) scheduleRetry(ctx context.Context, instance *workflowLogicInstance, logics []multiactionTuple, idx int, d time.Duration) error {

	var err error

	logics[idx].Attempts++
	logics[idx].ID = ""

	var data []byte
	data, err = json.Marshal(logics)
	if err != nil {
		return NewInternalError(err)
	}

	err = instance.Save(ctx, data)
	if err != nil {
		return err
	}

	r := &foreachStateLogicRetry{
		Idx:    idx,
		Logics: logics,
	}
	data = r.Marshal()

	t := time.Now().Add(d)

	err = instance.engine.scheduleRetry(instance.id, sl.ID(), instance.step, t, data)
	if err != nil {
		return err
	}

	return nil

}
