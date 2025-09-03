package transpiler

import (
	"context"

	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"gorm.io/gorm"
)

type Compiler struct {
	db         *gorm.DB
	transpiler *Transpiler
}

type TypescriptFlow struct {
	Script, Mapping string
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

func (c *Compiler) Compile(ctx context.Context, namespace, path string) (*TypescriptFlow, error) {
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

	return &TypescriptFlow{
		Script:  script,
		Mapping: mapping,
	}, nil
}

func TestCompile(script string) error {
	t, err := NewTranspiler()
	if err != nil {
		return err
	}

	_, _, err = t.Transpile(script, "dummy")

	return err
}
