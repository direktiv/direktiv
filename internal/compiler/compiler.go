package compiler

import (
	"context"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/internal/cluster/cache"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"gorm.io/gorm"
)

type Compiler struct {
	db    *gorm.DB
	cache cache.Cache
}

type CompileItem struct {
	tsScript         []byte
	path             string
	ValidationErrors []error

	script, mapping string
	config          *core.FlowConfig
}

func NewCompiler(db *gorm.DB, cache cache.Cache) (*Compiler, error) {
	return &Compiler{
		db:    db,
		cache: cache,
	}, nil
}

func (c *Compiler) getFile(ctx context.Context, namespace, path string) ([]byte, error) {
	f, err := filesql.NewStore(c.db).ForRoot(namespace).GetFile(ctx, path)
	if err != nil {
		return nil, err
	}

	return filesql.NewStore(c.db).ForFile(f).GetData(ctx)
}

func (c *Compiler) FetchScript(ctx context.Context, namespace, path string) (*core.TypescriptFlow, error) {
	// cacheKey := fmt.Sprintf("%s-%s-%s", namespace, "script", path)
	// flow, found := c.cache.Get(cacheKey)
	// if found {
	// 	return flow.(*core.TypescriptFlow), nil
	// }

	b, err := c.getFile(ctx, namespace, path)
	if err != nil {
		return nil, err
	}

	ci := &CompileItem{
		tsScript: b,
		path:     path,
	}

	err = ci.TranspileAndValidate()
	if err != nil {
		return nil, err
	}

	if len(ci.ValidationErrors) > 0 {
		errList := make([]string, len(ci.ValidationErrors))
		for i := range ci.ValidationErrors {
			errList[i] = ci.ValidationErrors[i].Error()
		}

		return nil, fmt.Errorf("%s", strings.Join(errList, ", "))
	}

	// c.cache.Set(cacheKey, obj)

	return ci.Config(), nil
}

func NewCompileItem(script []byte, path string) *CompileItem {
	return &CompileItem{
		tsScript:         script,
		path:             path,
		ValidationErrors: make([]error, 0),
	}
}

func (ci *CompileItem) Config() *core.TypescriptFlow {
	return &core.TypescriptFlow{
		Script:  ci.script,
		Mapping: ci.mapping,
		Config:  ci.config,
	}
}

func (ci *CompileItem) TranspileAndValidate() error {
	transpiler, err := NewTranspiler()
	if err != nil {
		return err
	}

	ci.script, ci.mapping, err = transpiler.Transpile(string(ci.tsScript), ci.path)

	return ci.validate()
}

func (ci *CompileItem) validate() error {
	pr, err := NewASTParser(ci.script, ci.mapping)
	if err != nil {
		return err
	}

	pr.ValidateTransitions()
	pr.ValidateFunctionCalls()

	config, err := pr.ValidateConfig()
	if err != nil {
		pr.Errors = append(pr.Errors, &ValidationError{
			Message: err.Error(),
			Line:    0,
			Column:  0,
		})
	}

	for i := range pr.Errors {
		ci.ValidationErrors = append(ci.ValidationErrors, pr.Errors[i])
	}

	ci.config = config

	return nil
}
