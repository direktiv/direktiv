// nolint
package service

import "io"

type client interface {
	createService(cfg *ServiceConfig) error
	updateService(cfg *ServiceConfig) error
	deleteService(id string) error
	listServices() ([]Status, error)
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
