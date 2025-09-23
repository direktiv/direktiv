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
	db         *gorm.DB
	transpiler *Transpiler
	cache      cache.Cache
}

func NewCompiler(db *gorm.DB, cache cache.Cache) (*Compiler, error) {
	transpiler, err := NewTranspiler()
	if err != nil {
		return nil, err
	}

	return &Compiler{
		db:         db,
		transpiler: transpiler,
		cache:      cache,
	}, nil
}

func (c *Compiler) FetchScript(ctx context.Context, namespace, path string) (*core.TypescriptFlow, error) {
	cacheKey := fmt.Sprintf("%s-%s-%s", namespace, "script", path)

	flow, found := c.cache.Get(cacheKey)
	if found {
		return flow.(*core.TypescriptFlow), nil
	}

	f, err := filesql.NewStore(c.db).ForRoot(namespace).GetFile(ctx, path)
	if err != nil {
		return nil, err
	}

	b, err := filesql.NewStore(c.db).ForFile(f).GetData(ctx)
	if err != nil {
		return nil, err
	}

	script, mapping, err := c.transpiler.Transpile(string(b), path)
	if err != nil {
		return nil, err
	}

	config, errs, err := ValidateScript(script)
	if err != nil {
		return nil, err
	}

	obj := &core.TypescriptFlow{
		Script:  script,
		Mapping: mapping,
		Config:  config,
	}

	if len(errs) > 0 {
		errList := make([]string, len(errs))
		for i := range errs {
			errList[i] = errs[i].Error()
		}
		return nil, fmt.Errorf("%s", strings.Join(errList, ", "))
	}
	// c.cache.Set(cacheKey, obj)

	return obj, nil
}

func ValidateScript(script string) (*core.FlowConfig, []error, error) {

	errors := make([]error, 0)

	t, err := NewTranspiler()
	if err != nil {
		return nil, errors, err
	}

	script, mapping, err := t.Transpile(script, "dummy")
	if err != nil {
		return nil, errors, err
	}

	pr, err := NewASTParser(script, mapping)
	if err != nil {
		return nil, errors, err
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
		errors = append(errors, pr.Errors[i])
	}

	return config, errors, nil
}
