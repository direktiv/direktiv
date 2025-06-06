package service

import (
	"io"

	"github.com/direktiv/direktiv/pkg/core"
)

// Services need a runtime that creates and schedule them, we use direktiv uses knative as a service runtime.
// runtimeClient interface implements imperative commands to manipulates services in the underlying runtime driver
// which is typically knative. Docker runtime driver is also implemented for demo purposes.
type runtimeClient interface {
	createService(sv *core.ServiceFileData) error
	updateService(sv *core.ServiceFileData) error
	deleteService(id string) error
	scaleService(id string, scale int32) error
	listServices() ([]status, error)
	cleanIdleServices(activeList []string) []error
	streamServiceLogs(id string, podID string) (io.ReadCloser, error)
	listServicePods(id string) (any, error)
}

type status interface {
	GetID() string
	GetValueHash() string
	GetConditions() any
}
