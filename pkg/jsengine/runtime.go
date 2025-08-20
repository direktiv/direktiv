package jsengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/dop251/goja"
)

const jsHiddenCode = `
function transition(funcName, state) {
	commitState(funcName, state)
	return funcName(state)
}

`

func (e *engine) execJSScript(script []byte, input string) (any, error) {
	vm := goja.New()
	vm.Set("print", jsPrint)
	vm.Set("commitState", jsCommitState)

	_, err := vm.RunString(jsHiddenCode + string(script))
	if err != nil {
		return nil, fmt.Errorf("run script: %s", err)
	}
	start, ok := goja.AssertFunction(vm.Get("start"))
	if !ok {
		return nil, errors.New("no start function")
	}

	var inputMap map[string]any
	err = json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal input: %s", err)
	}

	res, err := start(goja.Undefined(), vm.ToValue(inputMap))
	if err != nil {
		return nil, fmt.Errorf("invoke start: %s", err)
	}
	var result map[string]any
	if err := vm.ExportTo(res, &result); err != nil {
		return nil, fmt.Errorf("export output: %s", err)
	}

	return result, nil
}

func jsPrint(args ...any) {
	fmt.Print(args[0])
	if len(args) > 1 {
		for _, arg := range args[1:] {
			fmt.Print(" ")
			fmt.Print(arg)
		}
	}
	fmt.Println()
}

func jsCommitState(fn string, state any) {
	fmt.Printf("go: state committed fn:>%s< state:>%v<\n", parseJSFunctionName(fn), state)
}

func parseJSFunctionName(input string) string {
	re := regexp.MustCompile(`function\s+([a-zA-Z0-9_]+)\s*\(`)
	match := re.FindStringSubmatch(input)

	if len(match) > 1 {
		return (match[1])
	} else {
		return input
	}
}
