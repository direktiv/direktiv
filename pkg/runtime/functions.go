package runtime

import (
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/utils"
)

type FunctionCommand struct {
	functions *map[string]string

	rt *Runtime
}

func NewFunctionCommand(rt *Runtime, functions *map[string]string) *FunctionCommand {
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
		fid, err := compiler.GenerateFunctionID(in)
		if err != nil {
			throwRuntimeError(fc.rt.vm, DirektivFunctionErrorCode, err)
		}

		fn, ok := (*fc.functions)[fid]
		if !ok {
			fmt.Println("NOTHTHERER")
			throwRuntimeError(fc.rt.vm, DirektivFunctionErrorCode, fmt.Errorf("function does not exist"))
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

	runtime *Runtime
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
		throwRuntimeError(f.runtime.vm, DirektivFunctionErrorCode, err)
	}

	if f.url == "" {
		throwRuntimeError(f.runtime.vm, DirektivFunctionErrorCode, fmt.Errorf("function does not exist"))
	}

	slog.Debug("running function", slog.String("id", f.id))

	// response as file or json
	result := httpResultJSON
	if args.AsFile {
		result = httpResultFile
	}

	headers := make(map[string]string)
	headers[DirektivTempDir] = f.runtime.dirInfo().instanceDir
	headers[DirektivActionIDHeader] = f.runtime.id

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

	return f.runtime.HttpRequest(httpArgs)
}
