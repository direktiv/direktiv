package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/direktiv/direktiv/pkg/utils"
)

// func (rt *Runtime) flowFile(call goja.ConstructorCall) *goja.Object {
// 	if len(call.Arguments) != 1 {
// 		panic(rt.throwScriptError(runtimeErrorInitObject,
// 			"number of args not correct"))
// 	}

// 	f, err := NewFile(call.Arguments[0].Export(),
// 		rt.baseDir, rt.vm)
// 	if err != nil {
// 		panic(rt.throwScriptError(runtimeErrorInitObject,
// 			err.Error()))
// 	}

// 	return rt.vm.ToValue(f).ToObject(rt.vm)
// }

type FileCommand struct {
	rt *runtime.Runtime
}

func NewFileCommand(rt *runtime.Runtime) *FileCommand {
	return &FileCommand{
		rt: rt,
	}
}

func (fc FileCommand) GetName() string {
	return "getFile"
}

func (fc FileCommand) GetCommandFunction() interface{} {
	return func(in map[string]interface{}) *File {
		args, err := utils.DoubleMarshal[fileArgs](in)
		if err != nil {
			runtime.ThrowRuntimeError(fc.rt.VM, runtime.DirektivFileErrorCode, err)
		}

		if args.Scope == "" {
			args.Scope = fileScopeLocal
		}

		if args.Permission == 0 {
			args.Permission = 0777
		}

		perm := fmt.Sprintf("%v", args.Permission)
		o, err := strconv.ParseInt(perm, 8, 64)
		if err != nil {
			runtime.ThrowRuntimeError(fc.rt.VM, runtime.DirektivFileErrorCode, err)
		}
		args.Permission = int(o)

		f := &File{
			FileArgs: args,
			runtime:  fc.rt,
		}

		if !slices.Contains(allowedScopes, args.Scope) {
			runtime.ThrowRuntimeError(fc.rt.VM, runtime.DirektivFileErrorCode, fmt.Errorf("unknown scope %s", args.Scope))
		}

		if args.Name == "" {
			runtime.ThrowRuntimeError(fc.rt.VM, runtime.DirektivFileErrorCode, fmt.Errorf("filename empty"))
		}

		var prefixDir string
		if args.Scope == fileScopeLocal {
			prefixDir = fc.rt.DirInfo().InstanceDir
		} else if args.Scope == fileScopeShared {
			prefixDir = fc.rt.DirInfo().SharedDir
		}

		if prefixDir != "" {
			path := filepath.Join(prefixDir, args.Name)
			// if .. or something has been used
			if !strings.HasPrefix(path, prefixDir) {
				runtime.ThrowRuntimeError(fc.rt.VM, runtime.DirektivFileErrorCode, fmt.Errorf("illegal path for %s", args.Name))
			}
			f.RealPath = path
		}

		return f
	}
}

type fileArgs struct {
	Name       string `json:"name"`
	Permission int    `json:"permission"`
	Scope      string `json:"scope"`
}

type File struct {
	runtime  *runtime.Runtime
	FileArgs fileArgs `json:"fileArgs"`

	RealPath string
}

const (
	fileScopeLocal     = "local"
	fileScopeShared    = "shared"
	fileScopeWorfklow  = "workflow"
	fileScopeNamespace = "namespace"
)

var allowedScopes = []string{fileScopeLocal, fileScopeNamespace, fileScopeShared, fileScopeWorfklow}

// TODO: copy and move from local to shared and vice versa

func (f *File) Delete() {
	if f.RealPath != "" {
		err := os.Remove(f.RealPath)
		if err != nil {
			runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, err)
		}
	}
}

func (f *File) Data() string {
	switch f.FileArgs.Scope {
	case fileScopeShared:
		fallthrough
	case fileScopeLocal:
		b, err := os.ReadFile(f.RealPath)
		if err != nil {
			runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, err)
		}
		return string(b)
	default:
		runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, fmt.Errorf("not implemented"))
	}
	return ""
}

func (f *File) Name() string {
	return f.FileArgs.Name
}

func (f *File) Write(data string) {
	switch f.FileArgs.Scope {
	case fileScopeShared:
		fallthrough
	case fileScopeLocal:
		file, err := os.OpenFile(f.RealPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, fs.FileMode(f.FileArgs.Permission))
		if err != nil {
			runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, err)
		}
		defer file.Close()
		_, err = file.Write([]byte(data))
		if err != nil {
			runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, err)
		}
	default:
		runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, fmt.Errorf("not implemented"))
	}
}

func (f *File) Size() int {
	fi, err := os.Stat(f.RealPath)
	if err != nil {
		runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, err)
	}
	return int(fi.Size())
}

func (f *File) Base64() string {
	switch f.FileArgs.Scope {
	case fileScopeShared:
		fallthrough
	case fileScopeLocal:
		b, err := os.ReadFile(f.RealPath)
		if err != nil {
			runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, err)
		}
		return base64.StdEncoding.EncodeToString(b)
	default:
		runtime.ThrowRuntimeError(f.runtime.VM, runtime.DirektivFileErrorCode, fmt.Errorf("not implemented"))
	}
	return ""
}

func DoubleMarshal[T any](obj interface{}) (T, error) {
	var out T

	in, err := json.Marshal(obj)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(in, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}
