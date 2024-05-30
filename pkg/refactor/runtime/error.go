package runtime

import (
	"fmt"

	"github.com/dop251/goja"
)

type DirektivError struct {
	code, msg interface{}
}

func NewDirektivError(code, msg interface{}) *DirektivError {
	return &DirektivError{
		code: code,
		msg:  msg,
	}
}

func (de *DirektivError) Error() string {
	return fmt.Sprintf("%s - %s", de.code, de.msg)
}

func (de *DirektivError) Msg() string {
	return fmt.Sprintf("%s", de.msg)
}

func (de *DirektivError) Code() string {
	return fmt.Sprintf("%s", de.code)
}

func RuntimeDirektivError(call goja.ConstructorCall) *goja.Object {
	call.This.Set("code", call.Argument(0))
	call.This.Set("name", call.Argument(0))
	call.This.Set("msg", call.Argument(1))
	return nil
}

func throwRuntimeError(vm *goja.Runtime, code string, err error) {
	o, err := vm.New(vm.ToValue(RuntimeDirektivError), vm.ToValue(code),
		vm.ToValue(err.Error()))
	if err != nil {
		panic(err)
	}
	panic(o)
}
