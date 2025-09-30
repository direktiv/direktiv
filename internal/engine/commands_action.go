package engine

import (
	"encoding/json"
	"fmt"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/grafana/sobek"
)

func (cmds *Commands) action(call sobek.FunctionCall) sobek.Value {
	if len(call.Arguments) != 1 {
		panic(cmds.vm.ToValue("action definition needs configuration"))
	}

	imgConfig, ok := call.Argument(0).ToObject(cmds.vm).Export().(map[string]any)
	if !ok {
		panic(cmds.vm.ToValue("action image configuration has wrong type"))
	}

	// double marshal to get the struct
	j, err := json.Marshal(imgConfig)
	if err != nil {
		panic(cmds.vm.ToValue(fmt.Sprintf("action image configuration can not be converted: %s", err.Error())))
	}

	var actionConfig core.ActionConfig
	if err := json.Unmarshal(j, &actionConfig); err != nil {
		panic(cmds.vm.ToValue(fmt.Sprintf("action image configuration can not be converted: %s", err.Error())))
	}

	actionFunc := func(call sobek.FunctionCall) sobek.Value {

		if len(call.Arguments) == 0 {

		}
		var retValue any
		fmt.Printf("Arguments: %v\n", call.Arguments)

		// TODO: ad namespace services
		switch actionConfig.Type {
		case core.FlowActionScopeLocal:
			callLocal(actionConfig, "")
		case core.FlowActionScopeSubflow:
		default:
			panic(cmds.vm.ToValue(fmt.Sprintf("unknown action type '%s'", actionConfig.Type)))
		}

		return cmds.vm.ToValue(retValue)
	}

	return cmds.vm.ToValue(actionFunc)
}

func callLocal(config core.ActionConfig, payload any) error {

	return nil
}
