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
	streamServiceLogs(id string, podID string) (io.ReadCloser, error)
	listServicePods(id string) (any, error)
	killService(id string) error
}

type status interface {
	reconcileObject
	GetConditions() any
}
