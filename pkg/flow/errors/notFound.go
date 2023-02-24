package errors

import (
	"errors"
	"os"

	"github.com/direktiv/direktiv/pkg/flow/ent"
)

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

	if errors.Is(err, os.ErrNotExist) {
		return true
	}

	nferr := new(NotFoundError)
	return errors.As(err, &nferr)
}
