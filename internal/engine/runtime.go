package engine

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/grafana/sobek/parser"
)

func (e *Engine) execJSScript(instID uuid.UUID, script string, mappings string, fn string, input string) (any, error) {
	vm := sobek.New()
	vm.SetMaxCallStackSize(256)

	if mappings != "" {
		vm.SetParserOptions(parser.WithSourceMapLoader(func(path string) ([]byte, error) {
			return []byte(mappings), nil
		}))
	}

	// add commands
	InjectCommands(vm, instID)

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
	var result map[string]any
	if err := vm.ExportTo(res, &result); err != nil {
		return nil, fmt.Errorf("export output: %w", err)
	}

	return result, nil
}

func parseJSFunctionName(input string) string {
	re := regexp.MustCompile(`function\s+([a-zA-Z0-9_]+)\s*\(`)
	match := re.FindStringSubmatch(input)

	if len(match) > 1 {
		return match[1]
	} else {
		return input
	}
}
