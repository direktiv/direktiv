package direktiv

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"github.com/vorteil/direktiv/pkg/varstore"

	"github.com/vorteil/direktiv/pkg/model"
)

type getterStateLogic struct {
	state *model.GetterState
}

func initGetterStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	getter, ok := state.(*model.GetterState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(getterStateLogic)
	sl.state = getter
	return sl, nil

}

func (sl *getterStateLogic) Type() string {
	return model.StateTypeGetter.String()
}

func (sl *getterStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *getterStateLogic) Retries() *model.RetryDefinition {
	return sl.state.RetryDefinition()
}

func (sl *getterStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *getterStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *getterStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *getterStateLogic) LogJQ() string {
	return sl.state.Log
}

func (sl *getterStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	// get
	namespaceID := instance.namespace
	workflowID := instance.rec.Edges.Workflow.ID.String()
	instanceID := instance.id
	m := make(map[string]interface{})

	for _, v := range sl.state.Variables {

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

		var r varstore.VarReader
		r, err = instance.engine.server.variableStorage.Retrieve(ctx, v.Key, scope...)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		var data []byte
		data, err = ioutil.ReadAll(r)
		if err != nil {
			err = NewInternalError(err)
			return
		}

		var x interface{}
		if len(data) == 0 {
			x = nil
		} else {
			err = json.Unmarshal(data, &x)
			if err != nil {
				x = data
				err = nil
			}
		}

		m[v.Key] = x

	}

	err = instance.StoreData("var", m)
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
