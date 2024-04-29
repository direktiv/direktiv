package states

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeGetter, Getter)
}

type getterLogic struct {
	*model.GetterState
	Instance
}

func Getter(instance Instance, state model.State) (Logic, error) {
	getter, ok := state.(*model.GetterState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(getterLogic)
	sl.Instance = instance
	sl.GetterState = getter

	return sl, nil
}

func (logic *getterLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	err := scheduleOnce(logic, wakedata)
	if err != nil {
		return nil, err
	}

	var vars []VariableSelector
	var ptrs []string

	m := make(map[string]interface{})

	for idx, v := range logic.Variables {
		key := ""
		var selector VariableSelector

		x, err := jqOne(logic.GetInstanceData(), v.Key) //nolint:contextcheck
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

		if ok := util.MatchesVarRegex(key); !ok && v.Scope != util.VarScopeFileSystem {
			return nil, derrors.NewCatchableError(ErrCodeInvalidVariableKey, "variable key must match regex: %s (got: %s)", util.RegexPattern, key)
		}

		as := key
		if v.As != "" {
			as = v.As
		}

		selector.Key = key
		selector.Scope = v.Scope

		switch v.Scope {
		case "":
			selector.Scope = util.VarScopeInstance
			fallthrough

		case util.VarScopeInstance:
			fallthrough

		case util.VarScopeThread:
			fallthrough

		case util.VarScopeWorkflow:
			fallthrough

		case util.VarScopeFileSystem:
			fallthrough

		case util.VarScopeNamespace:
			vars = append(vars, selector)
			ptrs = append(ptrs, as)

		case util.VarScopeSystem:

			value, err := valueForSystem(key, logic.Instance)
			if err != nil {
				return nil, derrors.NewInternalError(err)
			}

			m[as] = value

		default:
			return nil, derrors.NewInternalError(errors.New("invalid scope"))
		}
	}

	results, err := logic.GetVariables(ctx, vars)
	if err != nil {
		return nil, err
	}

	for idx := range results {
		result := results[idx]
		as := ptrs[idx]

		var x interface{}

		x = nil

		if len(result.Data) != 0 {
			err = json.Unmarshal(result.Data, &x)
			if err != nil {
				x = result.Data
			}
		}

		m[as] = x
	}

	err = logic.StoreData("var", m)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}

func valueForSystem(key string, instance Instance) (interface{}, error) {
	var ret interface{}

	switch key {
	case "instance":
		ret = instance.GetInstanceID()
	case "uuid":
		ret = uuid.New().String()
	case "epoch":
		ret = time.Now().UTC().Unix()
	default:
		return nil, fmt.Errorf("unknown system key %s", key)
	}

	return ret, nil
}
