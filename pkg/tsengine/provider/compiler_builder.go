package provider

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
)

func BuildCompiler(ctx context.Context, provider FileGetter, namespace string, flowPath string) (tsservice.Compiler, error) {
	slog.Info("building flow", "flowPath", flowPath)

	b, err := provider.GetFileData(ctx, namespace, flowPath)
	if err != nil {
		return tsservice.Compiler{}, &FlowBuildError{flowPath: flowPath, err: err}
	}

	c, err := tsservice.New(flowPath, string(b))
	if err != nil {
		return tsservice.Compiler{}, &FlowBuildError{flowPath: flowPath, err: err}
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
