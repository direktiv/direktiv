package errors

import "fmt"

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
