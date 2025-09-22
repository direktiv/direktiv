package compiler

import (
	"context"
	"fmt"

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

	config, err := ValidateConfig(script, mapping)
	if err != nil {
		return nil, err
	}

	obj := &core.TypescriptFlow{
		Script:  script,
		Mapping: mapping,
		Config:  config,
	}

	c.cache.Set(cacheKey, obj)

	return obj, nil
}

func ValidateScript(script string) (*core.FlowConfig, error) {
	t, err := NewTranspiler()
	if err != nil {
		return nil, err
	}

	script, mapping, err := t.Transpile(script, "dummy")
	if err != nil {
		return nil, err
	}

	errors, err := ValidateTransitions(script, mapping)
	if err != nil {
		fmt.Println("HIER1")
		return nil, err
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors in script: %v", errors)
	}

	// err = ValidateBody(script, mapping)
	// if err != nil {
	// 	fmt.Println("HIER2")
	// 	return nil, err
	// }

	return ValidateConfig(script, mapping)
}
