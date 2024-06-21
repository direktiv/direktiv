package transpiler

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/dop251/goja"
	"github.com/thanhpk/randstr"
	"golang.org/x/exp/slog"
)

//go:embed ts-5.4.5.js
var TypescriptSource string

type Transpiler struct {
	vm  *goja.Runtime
	prg *goja.Program
	fn  string
}

func NewTranspiler() (Transpiler, error) {
	fn := randstr.String(8, "abcdefghijklmnopqrstuvwABCDEFGHIJKLMNOPQRSTUVWXYZ")
	vm := goja.New()

	err := vm.Set(fn, func(call goja.FunctionCall) goja.Value {
		bs, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			slog.Error("Base64 decoding error in transpiler function", "error", err)
			return goja.Undefined()
		}

		return vm.ToValue(string(bs))
	})
	if err != nil {
		return Transpiler{}, fmt.Errorf("failed to set transpiler function: %w", err)
	}

	program, err := goja.Compile("", TypescriptSource, true)
	if err != nil {
		slog.Error("Typescript compilation error", "error", err)
		return Transpiler{}, fmt.Errorf("failed to compile Typescript: %w", err)
	}

	_, err = vm.RunProgram(program)
	if err != nil {
		slog.Error("Typescript execution error", "error", err)
		return Transpiler{}, fmt.Errorf("failed to run Typescript: %w", err)
	}

	return Transpiler{vm: vm, prg: program, fn: fn}, nil
}

func (t *Transpiler) Transpile(script string) (string, error) {
	s := fmt.Sprintf("ts.transpile(%s('%s'), {}, /*fileName*/ undefined, /*diagnostics*/ undefined, /*moduleName*/ \"default\")",
		t.fn, base64.StdEncoding.EncodeToString([]byte(script)))

	value, err := t.vm.RunString(s)
	if err != nil {
		slog.Error("Transpilation script execution error", "error", err)
		return "", fmt.Errorf("failed to run transpilation script: %w", err)
	}

	s, ok := value.Export().(string)
	if !ok {
		slog.Error("Transpilation result type error")
		return "", fmt.Errorf("unexpected error during transpile: result is not a string")
	}

	_, err = goja.Compile("", s, true)
	if err != nil {
		slog.Error("Transpiled code compilation error", "error", err)
		return "", fmt.Errorf("failed to compile transpiled code: %w", err)
	}

	return s, nil
}
