package flow

import (
	"context"
	"encoding/json"
	"errors"
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

func (sl *setterStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
}

func (sl *setterStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *setterStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *setterStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *setterStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *setterStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	// set

	for _, v := range sl.state.Variables {

		var x interface{}

		x, err = jqOne(im.data, v.Value)
		if err != nil {
			return nil, err
		}

		var data []byte

		data, err = json.Marshal(x)
		if err != nil {
			return nil, NewInternalError(err)
		}

		hash := checksum(data)

		tx, err := engine.db.Tx(ctx)
		if err != nil {
			return nil, err
		}
		defer rollback(tx)

		vdatac := tx.VarData
		vrefc := tx.VarRef

		vdata, err := vdatac.Create().SetSize(len(data)).SetHash(hash).SetData(data).Save(ctx)
		if err != nil {
			return nil, err
		}

		switch v.Scope {

		case "":

			fallthrough

		case "instance":

			_, err = vrefc.Create().SetVardata(vdata).SetInstance(im.in).SetName(v.Key).Save(ctx)

		case "workflow":

			wf, err := engine.InstanceWorkflow(ctx, im)
			if err != nil {
				return nil, err
			}

			_, err = vrefc.Create().SetVardata(vdata).SetWorkflow(wf).SetName(v.Key).Save(ctx)

		case "namespace":

			ns, err := engine.InstanceNamespace(ctx, im)
			if err != nil {
				return nil, err
			}

			_, err = vrefc.Create().SetVardata(vdata).SetNamespace(ns).SetName(v.Key).Save(ctx)

		default:

			return nil, NewInternalError(errors.New("invalid scope"))

		}

		if err != nil {
			return nil, err
		}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
