package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/grafana/sobek"
)

func ExecScript(instID uuid.UUID, script string, mappings string, fn string,
	input string, metadata map[string]string,
) ([]byte, error) {
	// add commands

	rt := New(instID, metadata, mappings)

	_, err := rt.vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("run script: %w", err)
	}
	start, ok := sobek.AssertFunction(rt.vm.Get(fn))
	if !ok {
		return nil, fmt.Errorf("start function '%s' does not exist", fn)
	}

	var inputMap any
	err = json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal input: %w", err)
	}

	ret, err := start(sobek.Undefined(), rt.vm.ToValue(inputMap))
	if err != nil {
		return nil, fmt.Errorf("invoke start: %w", err)
	}
	var output any
	if err := rt.vm.ExportTo(ret, &output); err != nil {
		return nil, fmt.Errorf("export output: %w", err)
	}
	b, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	return b, nil
}
