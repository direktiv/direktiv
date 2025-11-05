package compiler

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"path/filepath"

	"github.com/grafana/sobek"
	"github.com/thanhpk/randstr"
)

//go:embed ts-5.9.2.js
var TypescriptSource string

type Transpiler struct {
	vm  *sobek.Runtime
	prg *sobek.Program
	fn  string
}

func NewTranspiler() (*Transpiler, error) {
	fn := randstr.String(8, "abcdefghijklmnopqrstuvwABCDEFGHIJKLMNOPQRSTUVWXYZ")
	vm := sobek.New()

	err := vm.Set(fn, func(call sobek.FunctionCall) sobek.Value {
		bs, _ := base64.StdEncoding.DecodeString(call.Argument(0).String())
		return vm.ToValue(string(bs))
	})
	if err != nil {
		return nil, err
	}

	program, err := sobek.Compile("", TypescriptSource, true)
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

func (t *Transpiler) Transpile(script, name string) (string, string, error) {
	// script = script + transitionCode

	s := fmt.Sprintf("ts.transpileModule(%s('%s'), { compilerOptions: { sourceMap: true }, fileName: \"%s\", moduleName: \"default\", reportDiagnostics: false })",
		t.fn, base64.StdEncoding.EncodeToString([]byte(script)), filepath.Base(name))

	value, err := t.vm.RunString(s)
	if err != nil {
		return "", "", err
	}

	// returns mapping and source file
	g := value.Export().(map[string]any) //nolint:forcetypeassert

	scriptOut, ok := g["outputText"].(string)
	if !ok {
		return "", "", fmt.Errorf("can not compile to js")
	}

	mappingOut, ok := g["sourceMapText"].(string)
	if !ok {
		return "", "", fmt.Errorf("can not generate mapping file")
	}

	return scriptOut, mappingOut, nil
}
