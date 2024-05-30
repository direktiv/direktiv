package errors

import (
	"errors"
	"os"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/filestore"
)

type NotFoundError struct {
	Label string
}

func (err *NotFoundError) Error() string {
	return err.Label
}

func IsNotFound(err error) bool {
	if errors.Is(err, filestore.ErrNotFound) || errors.Is(err, datastore.ErrNotFound) {
		return true
	}

	if errors.Is(err, os.ErrNotExist) {
		return true
	}

	nferr := new(NotFoundError)

	return errors.As(err, &nferr)
}
