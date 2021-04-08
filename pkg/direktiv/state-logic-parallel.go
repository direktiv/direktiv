package direktiv

import (
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

func (sl *parallelStateLogic) dispatchActions(ctx context.Context, instance *workflowLogicInstance, savedata []byte) error {

	var err error

	logics := make([]multiactionTuple, 0)

	if len(savedata) != 0 {
		return NewInternalError(errors.New("got unexpected savedata"))
	}

	if len(sl.state.Actions) > maxParallelActions {
		return NewUncatchableError("direktiv.limits.parallel", "instance aborted for exceeding the maximum number of parallel actions (%d)", maxParallelActions)
	}

	for _, action := range sl.state.Actions {

		var input interface{}

		input, err = jqObject(instance.data, ".")
		if err != nil {
			return err
		}

		m, ok := input.(map[string]interface{})
		if !ok {
			return NewInternalError(errors.New("invalid state data"))
		}

		m, err = addSecrets(ctx, instance, m, action.Secrets...)
		if err != nil {
			return err
		}

		input, err = jqObject(m, action.Input)
		if err != nil {
			return err
		}

		var inputData []byte

		inputData, err = json.Marshal(input)
		if err != nil {
			return NewInternalError(err)
		}

		if action.Function != "" {

			// container

			uid := ksuid.New()
			logics = append(logics, multiactionTuple{
				ID:   uid.String(),
				Type: "isolate",
			})

			var fn *model.FunctionDefinition
			fn, err = sl.workflow.GetFunction(action.Function)
			if err != nil {
				return NewInternalError(err)
			}

			ar := new(isolateRequest)
			ar.ActionID = uid.String()
			ar.Workflow.InstanceID = instance.id
			ar.Workflow.Namespace = instance.namespace
			ar.Workflow.State = sl.state.GetID()
			ar.Workflow.Step = instance.step
			ar.Workflow.Name = instance.wf.Name

			// TODO: timeout
			ar.Container.Data = inputData
			ar.Container.Image = fn.Image
			ar.Container.Cmd = fn.Cmd
			ar.Container.Size = fn.Size

			err = instance.engine.doActionRequest(ctx, ar)
			if err != nil {
				return err
			}

		} else {

			// subflow

			caller := new(subflowCaller)
			caller.InstanceID = instance.id
			caller.State = sl.state.GetID()
			caller.Step = instance.step

			var subflowID string

			subflowID, err = instance.engine.subflowInvoke(caller, instance.rec.InvokedBy, instance.namespace, action.Workflow, inputData)
			if err != nil {
				return err
			}

			logics = append(logics, multiactionTuple{
				ID:   subflowID,
				Type: "subflow",
			})

		}

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

func (sl *parallelStateLogic) LogJQ() string {
	return sl.state.Log
}

func (sl *parallelStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {
		err = sl.dispatchActions(ctx, instance, savedata)
		return
	}

	results := new(actionResultPayload)
	err = json.Unmarshal(wakedata, results)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	var logics []multiactionTuple
	err = json.Unmarshal(savedata, &logics)
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
			logics[i].Complete = true
			lid.Complete = true
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

		if completed == len(logics) {
			ready = true
		}

		if results.ErrorCode != "" {
			err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
			return
		}

		if results.ErrorMessage != "" {
			err = NewInternalError(errors.New(results.ErrorMessage))
			return
		}

	case model.BranchModeOr:

		if results.ErrorCode != "" {
			instance.Log("Branch %d failed with error '%s': %s", idx, results.ErrorCode, results.ErrorMessage)
		} else if results.ErrorMessage != "" {
			instance.Log("Branch %d failed with an internal error: %s", idx, results.ErrorMessage)
		} else {
			ready = true
		}

		if completed == len(logics) {
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
