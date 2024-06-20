package provider

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
)

// FunctionBuilder builds functions from the provided function provider.
type FunctionBuilder struct {
	functions map[string]tsservice.Function
	provider  FunctionProvider
}

// NewFunctionBuilder creates a new FunctionBuilder.
func NewFunctionBuilder(provider FunctionProvider, fi tsservice.FlowInformation) *FunctionBuilder {
	return &FunctionBuilder{
		functions: fi.Functions,
		provider:  provider,
	}
}

// Build retrieves and builds functions using the function provider.
func (b *FunctionBuilder) Build(ctx context.Context) map[string]string {
	functionsRet := make(map[string]string)
	for _, f := range b.functions {
		functionID := f.GetID()
		if f.Image != "" { // Only consider functions with non-empty images
			value, err := b.provider.GetFunction(ctx, functionID)
			if err != nil {
				slog.Error("failed to get function", slog.String("functionID", functionID), slog.Any("error", err))
				continue
			}
			functionsRet[functionID] = value
			slog.Info("adding function", slog.String("function", functionID))
		}
	}
	return functionsRet
}
