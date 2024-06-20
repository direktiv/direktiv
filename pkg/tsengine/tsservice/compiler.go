package tsservice

import (
	"github.com/direktiv/direktiv/pkg/tsengine/transpiler"
	"github.com/dop251/goja"
	"github.com/dop251/goja/ast"
)

type Compiler struct {
	// config     Config
	Path       string
	JavaScript string
	Program    *goja.Program

	ast *ast.Program
}

func New(path, typeScript string) (*Compiler, error) {
	tt, err := transpiler.NewTranspiler()
	if err != nil {
		return nil, err
	}

	// make javascript from typescript
	js, err := tt.Transpile(typeScript)
	if err != nil {
		return nil, err
	}

	// check if it is parsable
	ast, err := goja.Parse(path, js)
	if err != nil {
		return nil, err
	}

	// checks if there are function calls in global
	err = validateBodyFunctions(ast)
	if err != nil {
		return nil, err
	}

	// pre compile
	prg, err := goja.Compile(path, js, true)
	if err != nil {
		return nil, err
	}

	return &Compiler{
		Path:       path,
		JavaScript: js,
		ast:        ast,
		Program:    prg,
	}, err
}
