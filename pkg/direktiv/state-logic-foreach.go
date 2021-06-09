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

func (sl *foreachStateLogic) Retries() *model.RetryDefinition {
	return sl.state.RetryDefinition()
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

func (sl *foreachStateLogic) LogJQ() string {
	return sl.state.Log
}

func (sl *foreachStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(wakedata) == 0 {

		// first part

		logics := make([]multiactionTuple, 0)

		if len(savedata) != 0 {
			err = NewInternalError(errors.New("got unexpected savedata"))
			return
		}

		var array []interface{}
		array, err = jq(instance.data, sl.state.Array)
		if err != nil {
			return
		}

		instance.Log("Generated %d objects to loop over.", len(array))

		if len(array) > maxParallelActions {
			err = NewUncatchableError("direktiv.limits.parallel", "instance aborted for exceeding the maximum number of parallel actions (%d)", maxParallelActions)
			return
		}

		action := sl.state.Action

		for _, inputSource := range array {

			var inputData []byte
			inputData, err = generateActionInput(ctx, instance, inputSource, sl.state.Action)
			if err != nil {
				return
			}

			if action.Function != "" {

				// container

				uid := ksuid.New()
				logics = append(logics, multiactionTuple{
					ID:   uid.String(),
					Type: "isolate",
				})

				var fn *model.FunctionDefinition
				fn, err = sl.workflow.GetFunction(sl.state.Action.Function)
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

				// TODO: log subflow instance IDs

				subflowID, err = instance.engine.subflowInvoke(ctx, caller, instance.rec.InvokedBy, instance.namespace, action.Workflow, inputData)
				if err != nil {
					return
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
			err = NewInternalError(err)
			return
		}

		err = instance.Save(ctx, data)
		if err != nil {
			return
		}

		return

	}

	// second part

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

	instance.Log("Action returned. (%d/%d)", completed, len(logics))

	if results.ErrorCode != "" {
		instance.Log("Action returned catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)
		err = NewCatchableError(results.ErrorCode, results.ErrorMessage)
		return
	}

	if results.ErrorMessage != "" {
		instance.Log("Action crashed due to an internal error.")
		err = NewInternalError(errors.New(results.ErrorMessage))
		return
	}

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
