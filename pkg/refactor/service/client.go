// nolint
package service

import (
	"io"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type client interface {
	createService(cfg *core.ServiceConfig) error
	updateService(cfg *core.ServiceConfig) error
	deleteService(id string) error
	listServices() ([]status, error)
	streamServiceLogs(id string, podNumber int) (io.ReadCloser, error)
}

type status interface {
	reconcileObject
	GetConditions() any
}
