package flow

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCodeInternal               = "direktiv.internal.error"
	ErrCodeWorkflowUnparsable     = "direktiv.workflow.unparsable"
	ErrCodeMultipleErrors         = "direktiv.workflow.multipleErrors"
	ErrCodeCancelledByParent      = "direktiv.cancels.parent"
	ErrCodeSoftTimeout            = "direktiv.cancels.timeout.soft"
	ErrCodeHardTimeout            = "direktiv.cancels.timeout.hard"
	ErrCodeJQBadQuery             = "direktiv.jq.badCommand"
	ErrCodeJQNotObject            = "direktiv.jq.notObject"
	ErrCodeAllBranchesFailed      = "direktiv.parallel.allFailed"
	ErrCodeFailedSchemaValidation = "direktiv.schema.failed"
)

var (
	ErrNotDir      = errors.New("not a directory")
	ErrNotWorkflow = errors.New("not a workflow")
)

type CatchableError struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
}

func NewCatchableError(code string, msg string, a ...interface{}) *CatchableError {
	return &CatchableError{
		Code:    code,
		Message: fmt.Sprintf(msg, a...),
	}
}

func (err *CatchableError) Error() string {
	return err.Message
}

func WrapCatchableError(msg string, err error) error {

	if cerr, ok := err.(*CatchableError); ok {
		return &CatchableError{
			Code:    cerr.Code,
			Message: fmt.Sprintf(msg, err),
		}
	} else {
		return err
	}

}

type UncatchableError struct {
	Code    string
	Message string
}

func NewUncatchableError(code, msg string, a ...interface{}) *UncatchableError {
	return &UncatchableError{
		Code:    code,
		Message: fmt.Sprintf(msg, a...),
	}
}

func (err *UncatchableError) Error() string {
	return err.Message
}

type InternalError struct {
	Err      error
	Function string
	File     string
	Line     int
}

func NewInternalError(err error) error {
	if _, ok := err.(*InternalError); ok {
		return err
	}
	if _, ok := err.(*UncatchableError); ok {
		return err
	}
	if _, ok := err.(*CatchableError); ok {
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

type NotFoundError struct {
	Label string
}

func (err *NotFoundError) Error() string {
	return err.Label
}

func IsNotFound(err error) bool {
	if ent.IsNotFound(err) {
		return true
	}
	_, ok := err.(*NotFoundError)
	return ok
}

func translateError(err error) error {

	if IsNotFound(err) {
		err = status.Error(codes.NotFound, strings.TrimPrefix(err.Error(), "ent: "))
		return err
	}

	if _, ok := err.(*ent.ConstraintError); ok {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			err = status.Error(codes.AlreadyExists, "resource already exists")
			return err
		}
	}

	return err

}
