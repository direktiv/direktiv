package states

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeSetter, Setter)
}

type setterLogic struct {
	*model.SetterState
	Instance
}

func Setter(instance Instance, state model.State) (Logic, error) {
	setter, ok := state.(*model.SetterState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(setterLogic)
	sl.Instance = instance
	sl.SetterState = setter

	return sl, nil
}

//nolint:gocognit
func (logic *setterLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	err := scheduleOnce(logic, wakedata)
	if err != nil {
		return nil, err
	}

	setters := make([]VariableSetter, 0)

	for idx, v := range logic.Variables {
		var x interface{}
		key := ""
		mimeType := ""

		x, err = jqOne(logic.GetInstanceData(), v.Key) //nolint:contextcheck
		if err != nil {
			return nil, err
		}

		if x != nil {
			if str, ok := x.(string); ok {
				key = str
			}
		}

		if key == "" {
			return nil, derrors.NewCatchableError(ErrCodeJQNotString, "failed to evaluate key as a string for variable at index [%v]", idx)
		}

		if ok := util.MatchesVarRegex(key); !ok {
			return nil, derrors.NewCatchableError(ErrCodeInvalidVariableKey, "variable key must match regex: %s (got: %s)", util.RegexPattern, key)
		}

		if v.MimeType != nil {
			x, err = jqOne(logic.GetInstanceData(), v.MimeType) //nolint:contextcheck
			if err != nil {
				return nil, err
			}

			if x != nil {
				if str, ok := x.(string); ok {
					mimeType = str
				}
			}
		}

		fmt.Println("A", v.Value, reflect.TypeOf(v.Value), v.Value == nil)

		x, err = jqOne(logic.GetInstanceData(), v.Value) //nolint:contextcheck
		if err != nil {
			fmt.Println("B", err)
			return nil, err
		}

		fmt.Println("C")

		var data []byte

		if encodedData, ok := x.(string); ok && v.MimeType == "application/octet-stream" {
			decodedData, decodeErr := b64.StdEncoding.DecodeString(encodedData)
			if decodeErr != nil {
				return nil, derrors.NewInternalError(fmt.Errorf("could not decode variable '%s' base64 string %w", v.Key, err))
			}
			data = decodedData
		} else if v.MimeType == "text/plain; charset=utf-8" || v.MimeType == "text/plain" {
			data = []byte(fmt.Sprint(x))
		} else {
			data, err = json.Marshal(x)
			if err != nil {
				return nil, derrors.NewInternalError(err)
			}
		}

		setters = append(setters, VariableSetter{
			Scope:    v.Scope,
			Key:      key,
			MIMEType: mimeType,
			Data:     data,
		})
	}

	err = logic.SetVariables(ctx, setters)
	if err != nil {
		return nil, err
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
