package transpiler

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/dop251/goja"
	"github.com/thanhpk/randstr"
)

//go:embed ts-5.4.5.js
var TypescriptSource string

type Transpiler struct {
	vm  *goja.Runtime
	prg *goja.Program
	fn  string
}

func NewTranspiler() (*Transpiler, error) {

	fn := randstr.String(8, "abcdefghijklmnopqrstuvwABCDEFGHIJKLMNOPQRSTUVWXYZ")
	vm := goja.New()

	vm.Set(fn, func(call goja.FunctionCall) goja.Value {
		bs, _ := base64.StdEncoding.DecodeString(call.Argument(0).String())
		return vm.ToValue(string(bs))
	})

	program, err := goja.Compile("", TypescriptSource, true)
	if err != nil {
		return nil, err
	}

	_, err = vm.RunProgram(program)
	if err != nil {
		return nil, err
	}

	return &Transpiler{
		vm:  vm,
		prg: program,
		fn:  fn,
	}, nil

}

func (t *Transpiler) Transpile(script string) (string, error) {

	s := fmt.Sprintf("ts.transpile(%s('%s'), {}, /*fileName*/ undefined, /*diagnostics*/ undefined, /*moduleName*/ \"default\")",
		t.fn, base64.StdEncoding.EncodeToString([]byte(script)))
	value, err := t.vm.RunString(s)
	if err != nil {
		return "", err
	}

	s, ok := value.Export().(string)
	if !ok {
		return "", fmt.Errorf("unexpected error during transpile")
	}

	_, err = goja.Compile("", s, true)
	if err != nil {
		return "", err
	}

	return s, nil
}
