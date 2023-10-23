package service

import (
	"io"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

// Services need a runtime that creates and schedule them, we use direktiv uses knative as a service runtime.
// runtimeClient interface implements imperative commands to manipulates services in the underlying runtime driver
// which is typically knative. Docker runtime driver is also implemented for demo purposes.
type runtimeClient interface {
	createService(cfg *core.ServiceConfig) error
	updateService(cfg *core.ServiceConfig) error
	deleteService(id string) error
	listServices() ([]status, error)
	streamServiceLogs(id string, podID string) (io.ReadCloser, error)
	listServicePods(id string) (any, error)
	killService(id string) error
	getServiceURL(id string) string
}

type status interface {
	reconcileObject
	GetConditions() any
}
