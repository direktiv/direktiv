package errors

import (
	"errors"
	"fmt"
	"runtime"
)

type InternalError struct {
	Err      error
	Function string
	File     string
	Line     int
}

func NewInternalError(err error) error {
	ierr := new(InternalError)
	uerr := new(UncatchableError)
	cerr := new(CatchableError)

	if errors.As(err, &ierr) {
		return err
	}
	if errors.As(err, &uerr) {
		return err
	}
	if errors.As(err, &cerr) {
		return err
	}
	fn, file, line, _ := runtime.Caller(1)

	return &InternalError{
		Err:      err,
		Function: runtime.FuncForPC(fn).Name(),
		File:     file,
		Line:     line,
	}
}

func NewInternalErrorWithDepth(err error, depth int) *InternalError {
	fn, file, line, _ := runtime.Caller(depth)
	return &InternalError{
		Err:      err,
		Function: runtime.FuncForPC(fn).Name(),
		File:     file,
		Line:     line,
	}
}

func (err *InternalError) Error() string {
	return fmt.Sprintf("%s (%s %s:%v)", err.Err, err.Function, err.File, err.Line)
}

func (err *InternalError) Unwrap() error {
	return err.Err
}
