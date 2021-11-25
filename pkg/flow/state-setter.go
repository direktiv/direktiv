package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	b64 "encoding/base64"

	"github.com/direktiv/direktiv/pkg/model"
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

		if encodedData, ok := x.(string); ok && v.MimeType == "application/octet-stream" {
			decodedData, decodeErr := b64.StdEncoding.DecodeString(encodedData)
			if decodeErr != nil {
				return nil, NewInternalError(fmt.Errorf("could not decode variable '%s' base64 string %w", v.Key, err))
			}
			data = decodedData
		} else if v.MimeType == "text/plain; charset=utf-8" || v.MimeType == "text/plain" {
			data = []byte(fmt.Sprint(x))
		} else {
			data, err = json.Marshal(x)
			if err != nil {
				return nil, NewInternalError(err)
			}
		}

		// tx, err := engine.db.Tx(ctx)
		// if err != nil {
		// 	return nil, err
		// }
		// defer rollback(tx)
		//
		// vdatac := tx.VarData
		// vrefc := tx.VarRef

		vdatac := engine.db.VarData
		vrefc := engine.db.VarRef

		var q varQuerier

		switch v.Scope {

		case "":

			fallthrough

		case "instance":
			q = im.in
		case "workflow":
			wf, err := engine.InstanceWorkflow(ctx, im)
			if err != nil {
				return nil, err
			}

			wf, err = engine.db.Workflow.Get(ctx, wf.ID)
			if err != nil {
				return nil, err
			}

			q = wf
		case "namespace":
			ns, err := engine.InstanceNamespace(ctx, im)
			if err != nil {
				return nil, err
			}

			ns, err = engine.db.Namespace.Get(ctx, ns.ID)
			if err != nil {
				return nil, err
			}

			q = ns
		default:
			return nil, NewInternalError(errors.New("invalid scope"))
		}

		_, _, err = engine.flow.SetVariable(ctx, vrefc, vdatac, q, v.Key, data, v.MimeType)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		// TODO: make this a transaction?
		// err = tx.Commit()
		// if err != nil {
		// 	return nil, err
		// }

	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
