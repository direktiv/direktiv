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

func (sl *parallelStateLogic) Deadline() time.Time {
	return deadlineFromString(sl.state.Timeout)
}

func (sl *parallelStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *parallelStateLogic) ID() string {
	return sl.state.ID
}

func (sl *parallelStateLogic) LivingChildren(savedata []byte) []stateChild {

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

func (sl *parallelStateLogic) dispatchAction(ctx context.Context, instance *workflowLogicInstance, action *model.ActionDefinition, attempt int) (logic multiactionTuple, err error) {

	var inputData []byte
	inputData, err = generateActionInput(ctx, instance, instance.data, action)
	if err != nil {
		return
	}

	if action.Function != "" {

		// container

		uid := ksuid.New()
		logic = multiactionTuple{
			ID:       uid.String(),
			Type:     "isolate",
			Attempts: attempt,
		}

		var fn *model.FunctionDefinition
		fn, err = sl.workflow.GetFunction(action.Function)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		ar := new(isolateRequest)
		ar.ActionID = uid.String()
		ar.Workflow.InstanceID = instance.id
		ar.Workflow.Namespace = instance.namespace
		ar.Workflow.State = sl.state.GetID()
		ar.Workflow.Step = instance.step
		ar.Workflow.Name = instance.wf.Name
		ar.Workflow.ID = instance.wf.ID

		// TODO: timeout
		ar.Container.Data = inputData
		ar.Container.Image = fn.Image
		ar.Container.Cmd = fn.Cmd
		ar.Container.Size = fn.Size
		ar.Container.Scale = fn.Scale

		ar.Container.ID = fn.ID
		ar.Container.Files = fn.Files

		err = instance.engine.doActionRequest(ctx, ar)
		if err != nil {
			return
		}

	} else {

		// subflow

		caller := new(subflowCaller)
		caller.InstanceID = instance.id
		caller.State = sl.state.GetID()
		caller.Step = instance.step

		var subflowID string

		subflowID, err = instance.engine.subflowInvoke(ctx, caller, instance.rec.InvokedBy, instance.namespace, action.Workflow, inputData)
		if err != nil {
			return
		}

		logic = multiactionTuple{
			ID:       subflowID,
			Type:     "subflow",
			Attempts: attempt,
		}

	}

	return

}

func (sl *parallelStateLogic) dispatchActions(ctx context.Context, instance *workflowLogicInstance, savedata []byte) error {

	var err error

	logics := make([]multiactionTuple, 0)

	if len(savedata) != 0 {
		return NewInternalError(errors.New("got unexpected savedata"))
	}

	if len(sl.state.Actions) > maxParallelActions {
		return NewUncatchableError("direktiv.limits.parallel", "instance aborted for exceeding the maximum number of parallel actions (%d)", maxParallelActions)
	}

	for i := range sl.state.Actions {

		action := &sl.state.Actions[i]

		var logic multiactionTuple
		logic, err = sl.dispatchAction(ctx, instance, action, 0)
		if err != nil {
			return err
		}

		logics = append(logics, logic)

	}

	var data []byte
	data, err = json.Marshal(logics)
	if err != nil {
		return NewInternalError(err)
	}

	err = instance.Save(ctx, data)
	if err != nil {
		return err
	}

	return nil

}

func (sl *parallelStateLogic) doSpecific(ctx context.Context, instance *workflowLogicInstance, logics []multiactionTuple, idx int) (err error) {

	action := sl.state.Actions[idx]

	var logic multiactionTuple
	logic, err = sl.dispatchAction(ctx, instance, &action, logics[idx].Attempts)
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

func (sl *parallelStateLogic) LogJQ() string {
	return sl.state.Log
}

func (sl *parallelStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {
		err = sl.dispatchActions(ctx, instance, savedata)
		return
	}

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
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
		instance.Log("Retrying...")
		err = sl.doSpecific(ctx, instance, logics, retryData.Idx)
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
			instance.Log("Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)

			var d time.Duration
			d, err = preprocessRetry(sl.state.Actions[idx].Retries, logics[idx].Attempts, err)
			if err != nil {
				return
			}

			instance.Log("Scheduling retry attempt in: %v.", d)
			err = sl.scheduleRetry(ctx, instance, logics, idx, d)
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
		instance.Log("Action returned. (%d/%d)", completed, len(logics))
		if completed == len(logics) {
			ready = true
		}

	case model.BranchModeOr:

		if results.ErrorCode != "" {

			err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
			// instance.Log("Branch %d failed with error '%s': %s", idx, results.ErrorCode, results.ErrorMessage)
			instance.Log("Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
			var d time.Duration
			d, err = preprocessRetry(sl.state.Actions[idx].Retries, logics[idx].Attempts, err)
			if err == nil {
				err = sl.scheduleRetry(ctx, instance, logics, idx, d)
				return
			}

		} else if results.ErrorMessage != "" {
			instance.Log("Branch %d crashed due to an internal error: %s", idx, results.ErrorMessage)
			err = NewInternalError(errors.New(results.ErrorMessage))
			return
		} else {
			ready = true
		}

		logics[idx].Complete = true
		completed++
		instance.Log("Action returned. (%d/%d)", completed, len(logics))
		if !ready && completed == len(logics) {
			err = NewCatchableError(ErrCodeAllBranchesFailed, "all branches failed")
			return
		}

	default:
		err = NewInternalError(errors.New("unrecognized branch mode"))
		return
	}

	if !ready {
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

	var finalResults []interface{}
	for i := range logics {
		finalResults = append(finalResults, logics[i].Results)
	}

	err = instance.StoreData("return", finalResults)
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

func (sl *parallelStateLogic) scheduleRetry(ctx context.Context, instance *workflowLogicInstance, logics []multiactionTuple, idx int, d time.Duration) error {

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

	r := &parallelStateLogicRetry{
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
