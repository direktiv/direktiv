package errors

import (
	"errors"
	"fmt"
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
	cerr := new(CatchableError)

	//nolint:revive
	if errors.As(err, &cerr) {
		return &CatchableError{
			Code:    cerr.Code,
			Message: fmt.Sprintf(msg, err),
		}
	} else {
		return err
	}
}
