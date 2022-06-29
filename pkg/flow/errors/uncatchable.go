package errors

import "fmt"

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
