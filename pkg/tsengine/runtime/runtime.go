package runtime

import (
	"log/slog"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/google/uuid"
)

// Runtime represents a runtime environment for executing JavaScript code.
type Runtime struct {
	VM        *goja.Runtime
	ID        string
	baseDir   string
	JsonInput bool
}

func New(id uuid.UUID, baseDir string, jsonInput bool) (*Runtime, error) {
	slog.Debug("creating new runtime", slog.String("dir", baseDir), slog.String("instance", id.String()))

	vm := goja.New()
	vm.SetMaxCallStackSize(25)
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	vm.SetParserOptions(parser.WithDisableSourceMaps)

	rt := &Runtime{
		VM:        vm,
		ID:        id.String(),
		baseDir:   baseDir,
		JsonInput: jsonInput,
	}

	// err := vm.Set("log", commands.Log)
	// if err != nil {
	// 	return nil, err
	// }
	// err = vm.Set("sleep", commands.Sleep)
	// if err != nil {
	// 	return nil, err
	// }
	// err = vm.Set("atob", commands.Atob)
	// if err != nil {
	// 	return nil, err
	// }
	// err = vm.Set("btoa", commands.Btoa)
	// if err != nil {
	// 	return nil, err
	// }
	// err = vm.Set("toJSON", commands.ToJSON)
	// if err != nil {
	// 	return nil, err
	// }
	// err = vm.Set("fromJSON", commands.FromJSON)
	// if err != nil {
	// 	return nil, err
	// }
	// err = vm.Set("trim", commands.Trim)
	// if err != nil {
	// 	return nil, err
	// }

	return rt, nil
}

type Command interface {
	GetName() string
	GetCommandFunction() interface{}
}

func (rt *Runtime) WithCommand(command Command) error {
	return rt.VM.Set(command.GetName(), command.GetCommandFunction())
}

type dirInfo struct {
	SharedDir, InstanceDir string
}

func (rt *Runtime) DirInfo() *dirInfo {
	return &dirInfo{
		SharedDir:   filepath.Join(rt.baseDir, SharedDir),
		InstanceDir: filepath.Join(rt.baseDir, InstancesDir, rt.ID),
	}
}
