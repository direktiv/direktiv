package direktiv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/vorteil/direktiv/pkg/model"
)

type setterStateLogic struct {
	state *model.SetterState
}

func initSetterStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	setter, ok := state.(*model.SetterState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(setterStateLogic)
	sl.state = setter
	return sl, nil

}

func (sl *setterStateLogic) Type() string {
	return model.StateTypeSetter.String()
}

func (sl *setterStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *setterStateLogic) Retries() *model.RetryDefinition {
	return sl.state.RetryDefinition()
}

func (sl *setterStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *setterStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *setterStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *setterStateLogic) LogJQ() string {
	return sl.state.Log
}

func (sl *setterStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	// set
	namespaceID := instance.namespace
	workflowID := instance.rec.Edges.Workflow.ID.String()
	instanceID := instance.id

	for _, v := range sl.state.Variables {
		var x interface{}
		x, err = jqOne(instance.data, v.Value)
		if err != nil {
			return
		}

		scope := make([]string, 0)

		switch v.Scope {
		case "":
			fallthrough
		case "instance":
			scope = append(scope, namespaceID, workflowID, instanceID)
		case "workflow":
			scope = append(scope, namespaceID, workflowID)
		case "namespace":
			scope = append(scope, namespaceID)
		default:
			err = NewInternalError(errors.New("invalid scope"))
		}

		var data []byte
		data, err = json.Marshal(x)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		var w io.WriteCloser
		w, err = instance.engine.server.variableStorage.Store(ctx, v.Key, scope...)
		if err != nil {
			err = NewInternalError(err)
			return
		}
		defer w.Close()

		_, err = io.Copy(w, bytes.NewReader(data))
		if err != nil {
			err = NewInternalError(err)
			return
		}

		err = w.Close()
		if err != nil {
			err = NewInternalError(err)
			return
		}
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
