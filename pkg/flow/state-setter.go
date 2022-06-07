package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	b64 "encoding/base64"

	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
)

type setterStateLogic struct {
	*model.SetterState
}

func initSetterStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	setter, ok := state.(*model.SetterState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(setterStateLogic)
	sl.SetterState = setter

	return sl, nil
}

func (sl *setterStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
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

	tx, err := engine.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	vdatac := tx.VarData
	vrefc := tx.VarRef

	for idx, v := range sl.Variables {

		var x interface{}
		var key = ""
		var mimeType = ""

		x, err = jqOne(im.data, v.Key)
		if err != nil {
			return nil, err
		}

		if x != nil {
			if str, ok := x.(string); ok {
				key = str
			}
		}

		if key == "" {
			return nil, NewCatchableError(ErrCodeJQNotString, "failed to evaluate key as a string for variable at index [%v]", idx)
		}

		if ok := util.MatchesVarRegex(key); !ok {
			return nil, NewCatchableError(ErrCodeInvalidVariableKey, "variable key must match regex: %s (got: %s)", util.RegexPattern, key)
		}

		if v.MimeType != nil {
			x, err = jqOne(im.data, v.MimeType)
			if err != nil {
				return nil, err
			}

			if x != nil {
				if str, ok := x.(string); ok {
					mimeType = str
				}
			}

			// TODO: check that it matches a known mimetype?
		}

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

		// vdatac := engine.db.VarData
		// vrefc := engine.db.VarRef

		var q varQuerier

		var thread bool

		switch v.Scope {

		case "":

			fallthrough

		case "instance":
			q, err = tx.Instance.Get(ctx, im.in.ID)
			if err != nil {
				return nil, err
			}
			// q = im.in

		case "thread":
			q, err = tx.Instance.Get(ctx, im.in.ID)
			if err != nil {
				return nil, err
			}
			// q = im.in
			thread = true

		case "workflow":
			wf, err := engine.InstanceWorkflow(ctx, im)
			if err != nil {
				return nil, err
			}

			q, err = tx.Workflow.Get(ctx, wf.ID)
			if err != nil {
				return nil, err
			}

		case "namespace":
			ns, err := engine.InstanceNamespace(ctx, im)
			if err != nil {
				return nil, err
			}

			q, err = tx.Namespace.Get(ctx, ns.ID)
			if err != nil {
				return nil, err
			}

		default:
			return nil, NewInternalError(errors.New("invalid scope"))
		}

		_, _, err = engine.flow.SetVariable(ctx, vrefc, vdatac, q, key, data, mimeType, thread)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	transition = &stateTransition{
		Transform: sl.Transform,
		NextState: sl.Transition,
	}

	return

}
