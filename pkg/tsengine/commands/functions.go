package commands

import (
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/tsengine/runtime"
	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/direktiv/direktiv/pkg/utils"
)

type FunctionCommand struct {
	functions *map[string]string

	rt *runtime.Runtime
}

func NewFunctionCommand(rt *runtime.Runtime, functions *map[string]string) *FunctionCommand {
	return &FunctionCommand{
		rt:        rt,
		functions: functions,
	}
}

func (fc FunctionCommand) GetName() string {
	return "setupFunction"
}

func (fc FunctionCommand) GetCommandFunction() interface{} {
	return func(in map[string]interface{}) *Function {
		fid, err := tsservice.GenerateFunctionID(in)
		if err != nil {
			runtime.ThrowRuntimeError(fc.rt.VM, runtime.DirektivFunctionErrorCode, err)
		}

		fn, ok := (*fc.functions)[fid]
		if !ok {
			fmt.Println("NOTHTHERER")
			runtime.ThrowRuntimeError(fc.rt.VM, runtime.DirektivFunctionErrorCode, fmt.Errorf("function does not exist"))
		}

		return &Function{
			id:      fid,
			runtime: fc.rt,
			url:     fn,
		}
	}
}

type Function struct {
	id  string
	url string

	runtime *runtime.Runtime
}

type functionArgs struct {
	Input   interface{}
	File    *File
	Async   bool
	Retry   Retry
	AsFile  bool
	Timeout int
}

func (f *Function) Execute(in interface{}) interface{} {
	args, err := utils.DoubleMarshal[functionArgs](in)
	if err != nil {
		runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFunctionErrorCode, err)
	}

	if f.url == "" {
		runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFunctionErrorCode, fmt.Errorf("function does not exist"))
	}

	slog.Debug("running function", slog.String("id", f.id))

	// response as file or json
	result := httpResultJSON
	if args.AsFile {
		result = httpResultFile
	}

	headers := make(map[string]string)
	headers[runtime.DirektivTempDir] = f.runtime.DirInfo().InstanceDir
	headers[runtime.DirektivActionIDHeader] = f.runtime.ID

	httpArgs := HttpArgs{
		Method:  "POST",
		URL:     f.url,
		Header:  headers,
		Timeout: args.Timeout,
		Async:   args.Async,
		Retry:   args.Retry,
		Result:  result,
	}

	if args.File != nil {
		httpArgs.File = args.File
	} else {
		httpArgs.Input = args.Input
	}

	return NewRequestCommand(f.runtime).HttpRequest(httpArgs)
}
