package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/direktiv/direktiv/internal/cluster/cache"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"gorm.io/gorm"
)

type Compiler struct {
	db             *gorm.DB
	cache          cache.Cache[core.TypescriptFlow]
	secretsManager core.SecretsManager
}

type CompileItem struct {
	tsScript         []byte
	path             string
	ValidationErrors []error

	script, mapping string
	config          core.FlowConfig
}

func NewCompiler(db *gorm.DB, secretsManager core.SecretsManager, cache cache.Cache[core.TypescriptFlow]) (*Compiler, error) {
	return &Compiler{
		db:             db,
		cache:          cache,
		secretsManager: secretsManager,
	}, nil
}

func (c *Compiler) getFile(ctx context.Context, namespace, path string) ([]byte, error) {
	f, err := filesql.NewStore(c.db).ForRoot(namespace).GetFile(ctx, path)
	if err != nil {
		return nil, err
	}

	return filesql.NewStore(c.db).ForFile(f).GetData(ctx)
}

func (c *Compiler) FetchScript(ctx context.Context, namespace, path string, withSecrets bool) (core.TypescriptFlow, error) {
	cacheKey := fmt.Sprintf("%s-%s-%s", namespace, "script", path)
	flow, err := c.cache.Get(cacheKey, func(a ...any) (core.TypescriptFlow, error) {
		return c.genFlow(ctx, namespace, path)
	})

	if err != nil {
		slog.Error("cannot fetch sript during compile", slog.Any("error", err))
		return flow, err
	}

	secretMap := make(map[string][]byte)
	if withSecrets {
		for a := range flow.Config.Secrets {
			secret, err := c.secretsManager.Get(ctx, namespace, flow.Config.Secrets[a])
			if err != nil {
				slog.Error("cannot fetch secret during compile", slog.Any("error", err))
				return flow, err
			}

			secretMap[secret.Name] = secret.Data
		}
	}

	// store secrets as json map
	sm, _ := json.Marshal(secretMap)
	flow.Secrets = string(sm)

	return flow, nil
}

func (c *Compiler) genFlow(ctx context.Context, namespace, path string) (core.TypescriptFlow, error) {
	b, err := c.getFile(ctx, namespace, path)
	if err != nil {
		return core.TypescriptFlow{}, err
	}

	ci := &CompileItem{
		tsScript: b,
		path:     path,
	}

	err = ci.TranspileAndValidate()
	if err != nil {
		return core.TypescriptFlow{}, err
	}

	if len(ci.ValidationErrors) > 0 {
		errList := make([]string, len(ci.ValidationErrors))
		for i := range ci.ValidationErrors {
			errList[i] = ci.ValidationErrors[i].Error()
		}

		return core.TypescriptFlow{}, fmt.Errorf("%s", strings.Join(errList, ", "))
	}

	return ci.Config(), nil
}

func NewCompileItem(script []byte, path string) *CompileItem {
	return &CompileItem{
		tsScript:         script,
		path:             path,
		ValidationErrors: make([]error, 0),
	}
}

func (ci *CompileItem) Config() core.TypescriptFlow {
	return core.TypescriptFlow{
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
	if err != nil {
		return err
	}

	return ci.validate()
}

func (ci *CompileItem) validate() error {
	pr, err := NewASTParser(ci.script, ci.mapping)
	if err != nil {
		return err
	}

	err = pr.Parse()
	if err != nil {
		return err
	}

	ci.config = pr.FlowConfig
	ci.config.Actions = pr.Actions
	ci.config.Secrets = pr.allSecretNames

	for i := range pr.Errors {
		ci.ValidationErrors = append(ci.ValidationErrors, pr.Errors[i])
	}

	if pr.FirstStateFunc == "" {
		ci.ValidationErrors = append(ci.ValidationErrors, &ValidationError{
			Message:  "no state functions defined",
			Severity: SeverityError,
		})
	}

	return nil
}
