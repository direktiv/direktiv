package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/grafana/sobek"
)

type Script struct {
	InstID   uuid.UUID
	Text     string
	Mappings string
	Fn       string
	Input    string
	Metadata map[string]string
}

func ExecScript(sc *Script, cFinish CommitFinishStateFunc,
) error {
	// add commands

	rt := New(sc.InstID, sc.Metadata, sc.Mappings, cFinish)

	_, err := rt.vm.RunString(sc.Text)
	if err != nil {
		return fmt.Errorf("run script: %w", err)
	}
	start, ok := sobek.AssertFunction(rt.vm.Get(sc.Fn))
	if !ok {
		return fmt.Errorf("start function '%s' does not exist", sc.Fn)
	}

	var inputMap any
	err = json.Unmarshal([]byte(sc.Input), &inputMap)
	if err != nil {
		return fmt.Errorf("unmarshal input: %w", err)
	}

	_, err = start(sobek.Undefined(), rt.vm.ToValue(inputMap))
	if err != nil {
		return fmt.Errorf("invoke start: %w", err)
	}

	return nil
}
