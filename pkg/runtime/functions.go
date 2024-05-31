package runtime

import (
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/compiler"
)

type Function struct {
	id  string
	url string

	runtime *Runtime
}

func (rt *Runtime) setupFunction(in map[string]interface{}) *Function {
	fid, err := compiler.GenerateFunctionID(in)
	if err != nil {
		throwRuntimeError(rt.vm, DirektivFunctionErrorCode, err)
	}

	fn, ok := (*rt.Functions)[fid]
	if !ok {
		fmt.Println("NOTHTHERER")
		throwRuntimeError(rt.vm, DirektivFunctionErrorCode, fmt.Errorf("function does not exist"))
	}

	return &Function{
		id:      fid,
		runtime: rt,
		url:     fn,
	}
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
	args, err := compiler.DoubleMarshal[functionArgs](in)
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
	headers[DirektivTempDir] = f.runtime.baseDir
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
