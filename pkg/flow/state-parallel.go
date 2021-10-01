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

type parallelStateLogic struct {
	state    *model.ParallelState
	workflow *model.Workflow
}

func initParallelStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	parallel, ok := state.(*model.ParallelState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(parallelStateLogic)
	sl.state = parallel
	sl.workflow = wf

	return sl, nil

}

func (sl *parallelStateLogic) Type() string {
	return model.StateTypeParallel.String()
}

func (sl *parallelStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return deadlineFromString(ctx, engine, im, sl.state.Timeout)
}

func (sl *parallelStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *parallelStateLogic) ID() string {
	return sl.state.ID
}

func (sl *parallelStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {

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

func (sl *parallelStateLogic) dispatchAction(ctx context.Context, engine *engine, im *instanceMemory, action *model.ActionDefinition, attempt int) (logic multiactionTuple, err error) {

	var inputData []byte
	inputData, err = generateActionInput(ctx, engine, im, im.data, action)
	if err != nil {
		return
	}

	fn, err := sl.workflow.GetFunction(action.Function)
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

func (sl *parallelStateLogic) dispatchActions(ctx context.Context, engine *engine, im *instanceMemory) error {

	var err error

	logics := make([]multiactionTuple, 0)

	if im.GetMemory() != nil {
		return NewInternalError(errors.New("got unexpected savedata"))
	}

	for i := range sl.state.Actions {

		action := &sl.state.Actions[i]

		var logic multiactionTuple
		logic, err = sl.dispatchAction(ctx, engine, im, action, 0)
		if err != nil {
			return err
		}

		logics = append(logics, logic)

	}

	err = engine.SetMemory(ctx, im, logics)
	if err != nil {
		return err
	}

	return nil

}

func (sl *parallelStateLogic) doSpecific(ctx context.Context, engine *engine, im *instanceMemory, logics []multiactionTuple, idx int) (err error) {

	action := sl.state.Actions[idx]

	var logic multiactionTuple
	logic, err = sl.dispatchAction(ctx, engine, im, &action, logics[idx].Attempts)
	if err != nil {
		return
	}

	logics[idx] = logic

	err = engine.SetMemory(ctx, im, logics)
	if err != nil {
		return err
	}

	return

}

func (sl *parallelStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *parallelStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {
		err = sl.dispatchActions(ctx, engine, im)
		return
	}

	var logics []multiactionTuple
	err = im.UnmarshalMemory(&logics)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	// check for scheduled retry
	retryData := new(parallelStateLogicRetry)
	dec := json.NewDecoder(bytes.NewReader(wakedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(retryData)
	if err == nil {
		engine.logToInstance(ctx, time.Now(), im.in, "Retrying...")
		err = sl.doSpecific(ctx, engine, im, logics, retryData.Idx)
		return
	}

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

		if lid.ID == results.ActionID {
			found = true
			if lid.Complete {
				err = NewInternalError(fmt.Errorf("action '%s' already completed", lid.ID))
				return
			}
			idx = i
		}

		if lid.Complete {
			completed++
		}

	}

	if !found {
		err = NewInternalError(fmt.Errorf("action '%s' wasn't expected", results.ActionID))
		return
	}

	var x interface{}
	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	logics[idx].Results = x

	var ready bool
	switch sl.state.Mode {
	case model.BranchModeAnd:

		if results.ErrorCode != "" {

			err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
			engine.logToInstance(ctx, time.Now(), im.in, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)

			var d time.Duration
			d, err = preprocessRetry(sl.state.Actions[idx].Retries, logics[idx].Attempts, err)
			if err != nil {
				return
			}

			engine.logToInstance(ctx, time.Now(), im.in, "Scheduling retry attempt in: %v.", d)
			err = sl.scheduleRetry(ctx, engine, im, logics, idx, d)
			return

		}

		if results.ErrorMessage != "" {
			err = NewInternalError(errors.New(results.ErrorMessage))
			return
		}

		logics[idx].Complete = true
		if logics[idx].Complete {
			completed++
		}
		engine.logToInstance(ctx, time.Now(), im.in, "Action returned. (%d/%d)", completed, len(logics))
		if completed == len(logics) {
			ready = true
		}

	case model.BranchModeOr:

		if results.ErrorCode != "" {

			err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
			// instance.Log("Branch %d failed with error '%s': %s", idx, results.ErrorCode, results.ErrorMessage)
			engine.logToInstance(ctx, time.Now(), im.in, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
			var d time.Duration
			d, err = preprocessRetry(sl.state.Actions[idx].Retries, logics[idx].Attempts, err)
			if err == nil {
				err = sl.scheduleRetry(ctx, engine, im, logics, idx, d)
				return
			}

		} else if results.ErrorMessage != "" {
			engine.logToInstance(ctx, time.Now(), im.in, "Branch %d crashed due to an internal error: %s", idx, results.ErrorMessage)
			err = NewInternalError(errors.New(results.ErrorMessage))
			return
		} else {
			ready = true
		}

		logics[idx].Complete = true
		completed++
		engine.logToInstance(ctx, time.Now(), im.in, "Action returned. (%d/%d)", completed, len(logics))
		if !ready && completed == len(logics) {
			err = NewCatchableError(ErrCodeAllBranchesFailed, "all branches failed")
			return
		}

	default:
		err = NewInternalError(errors.New("unrecognized branch mode"))
		return
	}

	if !ready {
		err = engine.SetMemory(ctx, im, logics)
		return
	}

	var finalResults []interface{}
	for i := range logics {
		finalResults = append(finalResults, logics[i].Results)
	}

	err = im.StoreData("return", finalResults)
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

type parallelStateLogicRetry struct {
	Logics []multiactionTuple
	Idx    int
}

func (r *parallelStateLogicRetry) Marshal() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return data
}

func (sl *parallelStateLogic) scheduleRetry(ctx context.Context, engine *engine, im *instanceMemory, logics []multiactionTuple, idx int, d time.Duration) error {

	var err error

	logics[idx].Attempts++
	logics[idx].ID = ""

	err = engine.SetMemory(ctx, im, logics)
	if err != nil {
		return err
	}

	r := &parallelStateLogicRetry{
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
