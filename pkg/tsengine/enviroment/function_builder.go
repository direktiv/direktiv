package enviroment

import (
	"log/slog"
	"os"

	"github.com/direktiv/direktiv/pkg/compiler"
)

type FunctionBuilder struct {
	functions map[string]compiler.Function
	baseFS    string
}

func NewFunctionBuilder(fi compiler.FlowInformation, baseFS string) *FunctionBuilder {
	return &FunctionBuilder{
		functions: fi.Functions,
		baseFS:    baseFS,
	}
}

func (b *FunctionBuilder) Build() map[string]string {
	functionsRet := make(map[string]string)
	for _, f := range b.functions {
		functionID := f.GetID()
		value := os.Getenv(functionID)
		functionsRet[functionID] = value
		slog.Info("adding function", slog.String("function", functionID))
	}
	return functionsRet
}

// functions := make(map[string]string)
// for i := range fi.Functions {
// 	f := fi.Functions[i]
// 	// only do workflow functions
// 	if f.Image != "" {
// 		slog.Debug("adding function", slog.String("function", f.Image))
// 		functions[f.GetID()] = os.Getenv(f.GetID())
// 	}
// }
