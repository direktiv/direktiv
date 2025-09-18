package compiler

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"gorm.io/gorm"
)

const transitionCode = `
function transition(funcName, state) {
	commitState(funcName, state)
	return funcName(state)
}
`

type Compiler struct {
	db         *gorm.DB
	transpiler *Transpiler
}

func NewCompiler(db *gorm.DB) (*Compiler, error) {
	transpiler, err := NewTranspiler()
	if err != nil {
		return nil, err
	}

	return &Compiler{
		db:         db,
		transpiler: transpiler,
	}, nil
}

func (c *Compiler) FetchScript(ctx context.Context, namespace, path string) (*core.TypescriptFlow, error) {

	// TODO CACHING

	f, err := filesql.NewStore(c.db).ForRoot(namespace).GetFile(ctx, path)
	if err != nil {
		return nil, err
	}

	b, err := filesql.NewStore(c.db).ForFile(f).GetData(ctx)
	if err != nil {
		return nil, err
	}

	// add transition function
	appendScript := string(b) + transitionCode

	script, mapping, err := c.transpiler.Transpile(appendScript, path)
	if err != nil {
		return nil, err
	}

	config, err := ValidateConfig(script)

	return &core.TypescriptFlow{
		Script:  script,
		Mapping: mapping,
		Config:  config,
	}, nil
}

func ValidateScript(script string) (*core.FlowConfig, error) {
	t, err := NewTranspiler()
	if err != nil {
		return nil, err
	}

	script, _, err = t.Transpile(script, "dummy")

	errors, err := ValidateTransitions(script)
	if err != nil {
		return nil, err
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors in script: %v", errors)
	}

	err = ValidateBody(script)
	if err != nil {
		return nil, err
	}

	return ValidateConfig(script)
}
