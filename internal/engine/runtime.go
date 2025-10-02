package engine

import (
	"encoding/json"
	"fmt"

	"github.com/direktiv/direktiv/internal/engine/commands"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/grafana/sobek/parser"
)

func (e *Engine) execJSScript(instID uuid.UUID, script string, mappings string, fn string,
	input string, metadata map[string]string,
) (any, error) {
	vm := sobek.New()
	vm.SetMaxCallStackSize(256)

	if mappings != "" {
		vm.SetParserOptions(parser.WithSourceMapLoader(func(path string) ([]byte, error) {
			return []byte(mappings), nil
		}))
	}

	// add commands
	commands.InjectCommands(vm, instID, metadata)

	_, err := vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("run script: %w", err)
	}
	start, ok := sobek.AssertFunction(vm.Get(fn))
	if !ok {
		return nil, fmt.Errorf("start function '%s' does not exist", fn)
	}

	var inputMap map[string]any
	err = json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal input: %w", err)
	}

	res, err := start(sobek.Undefined(), vm.ToValue(inputMap))
	if err != nil {
		return nil, fmt.Errorf("invoke start: %w", err)
	}
	var result any
	if err := vm.ExportTo(res, &result); err != nil {
		return nil, fmt.Errorf("export output: %w", err)
	}

	return result, nil
}
