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

type ClientConfig struct {
	ServiceAccount string `yaml:"serviceAccount"`
	Namespace      string `yaml:"namespace"`
	IngressClass   string `yaml:"ingressClass"`

	Sidecar string `yaml:"sidecar"`

	MaxScale int    `yaml:"maxScale"`
	NetShape string `yaml:"netShape"`
}

type status interface {
	reconcileObject
	GetConditions() any
}
