package direktiv

import (
	"fmt"
	"runtime"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/schema"
)

var grpcErrInternal = status.Error(codes.Internal, "internal error")

var errorRegistry = map[string]codes.Code{
	"the workflow has already been updated": codes.AlreadyExists,
}

func grpcDatabaseError(err error, otype, oval string) error {

	if err == nil {
		return nil
	}

	if _, ok := err.(*UncatchableError); ok {
		return err
	}

	if _, ok := err.(*CatchableError); ok {
		return err
	}

	if _, ok := err.(*InternalError); ok {
		return err
	}

	if code, ok := errorRegistry[err.Error()]; ok {
		return status.Errorf(code, "%s '%s' "+err.Error(), otype, oval)
	}

	if ent.IsNotFound(err) {
		return status.Errorf(codes.NotFound, "%s '%s' does not exist", otype, oval)
	}

	if ent.IsConstraintError(err) {
		if strings.Contains(err.Error(), "duplicate key value") {
			return status.Errorf(codes.AlreadyExists, "%s '%s' already exists", otype, oval)
		}
	}

	if ent.IsValidationError(err) {
		if strings.HasSuffix(err.Error(), `"description": value is greater than the required length`) {
			return status.Errorf(codes.InvalidArgument, "description is greater than max length of %v bytes", schema.MaxLenDescription)
		}
	}

	// Handle GRPC errors
	if _, ok := err.(interface {
		GRPCStatus() *status.Status
	}); ok {
		return err
	}

	log.Errorf("%v", NewInternalErrorWithDepth(err, 2))

	err = grpcErrInternal

	return err

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

type CatchableError struct {
	Code    string
	Message string
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

type InternalError struct {
	Err      error
	Function string
	File     string
	Line     int
}

func NewInternalError(err error) *InternalError {
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

// ErrorType types of errors direktiv commands can return
type ErrorType int

const (
	DirektivError ErrorType = iota
	// Ent Errors
	ValidationError
	NotFoundError
	NotSingularError
	NotLoadedError
	ConstraintError
	// Other Errors
)

// CmdErrorResponse struct for responding when command has an error
type CmdErrorResponse struct {
	Error string    `json:"error"`
	Type  ErrorType `json:"type"`
}

// GetErrorType get Error Type from passed error
func GetErrorType(err error) ErrorType {
	if ent.IsValidationError(err) {
		return ValidationError
	}

	if ent.IsNotFound(err) {
		return NotFoundError
	}

	if ent.IsNotSingular(err) {
		return NotSingularError
	}

	if ent.IsNotLoaded(err) {
		return NotLoadedError
	}

	if ent.IsConstraintError(err) {
		return ConstraintError
	}

	return DirektivError
}
