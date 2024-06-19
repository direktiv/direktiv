package provider

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/tsengine/compiler"
)

func BuildCompiler(ctx context.Context, provider FileGetter, namespace string, flowPath string) (compiler.Compiler, error) {
	slog.Info("building flow", "flowPath", flowPath)

	b, err := provider.GetFileData(ctx, namespace, flowPath)
	if err != nil {
		return compiler.Compiler{}, &FlowBuildError{flowPath: flowPath, err: err}
	}

	c, err := compiler.New(flowPath, string(b))
	if err != nil {
		return compiler.Compiler{}, &FlowBuildError{flowPath: flowPath, err: err}
	}

	return *c, nil
}

type FlowBuildError struct {
	flowPath string
	err      error
}

func (e *FlowBuildError) Error() string {
	return fmt.Sprintf("Error building flow '%s': %v", e.flowPath, e.err)
}
