package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
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

func (sl *getterStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
}

func (sl *getterStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *getterStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *getterStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *getterStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *getterStateLogic) MetadataJQ() interface{} {
	return sl.state.Metadata
}

func (sl *getterStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	// get

	m := make(map[string]interface{})

	for _, v := range sl.state.Variables {

		var ref *ent.VarRef

		storeKey := v.Key
		if v.As != "" {
			storeKey = v.As
		}

		switch v.Scope {

		case "":

			fallthrough

		case "instance":

			ref, err = im.in.QueryVars().Where(entvar.NameEQ(v.Key), entvar.BehaviourIsNil()).WithVardata().Only(ctx)

		case "thread":

			ref, err = im.in.QueryVars().Where(entvar.NameEQ(v.Key), entvar.BehaviourEQ("thread")).WithVardata().Only(ctx)

		case "workflow":

			wf, err := engine.InstanceWorkflow(ctx, im)
			if err != nil {
				return nil, NewInternalError(err)
			}

			// NOTE: this hack seems to be necessary for some reason...
			wf, err = engine.db.Workflow.Get(ctx, wf.ID)
			if err != nil {
				return nil, NewInternalError(err)
			}

			ref, err = wf.QueryVars().Where(entvar.NameEQ(v.Key)).WithVardata().Only(ctx)

		case "namespace":

			ns, err := engine.InstanceNamespace(ctx, im)
			if err != nil {
				return nil, NewInternalError(err)
			}

			// NOTE: this hack seems to be necessary for some reason...
			ns, err = engine.db.Namespace.Get(ctx, ns.ID)
			if err != nil {
				return nil, NewInternalError(err)
			}

			ref, err = ns.QueryVars().Where(entvar.NameEQ(v.Key)).WithVardata().Only(ctx)

		case "system":

			value, err := valueForSystem(v.Key, im)
			if err != nil {
				return nil, NewInternalError(err)
			}

			m[storeKey] = value
			continue

		default:

			err = NewInternalError(errors.New("invalid scope"))

		}

		var data []byte

		if err != nil {
			if IsNotFound(err) {
				data = make([]byte, 0)
			} else {
				return nil, NewInternalError(err)
			}
		} else if ref == nil {
			data = make([]byte, 0)
		} else {
			if ref.Edges.Vardata == nil {
				err = &NotFoundError{
					Label: fmt.Sprintf("variable data not found"),
				}
				return nil, err
			}
			data = ref.Edges.Vardata.Data
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

		m[storeKey] = x

	}

	err = im.StoreData("var", m)
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

func valueForSystem(key string, im *instanceMemory) (interface{}, error) {

	var ret interface{}

	switch key {
	case "instance":
		ret = im.ID()
	case "uuid":
		ret = uuid.New().String()
	case "epoch":
		ret = time.Now().Unix()
	default:
		return nil, fmt.Errorf("unknown system key %s", key)
	}

	return ret, nil

}
