package enviroment

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/compiler"
)

func BuildCompiler(flowPath string, namespace string, provider FileGetter) (compiler.Compiler, error) {
	slog.Info("building flow", "flowPath", flowPath)

	b, err := provider.GetData(context.Background(), namespace, flowPath)
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
