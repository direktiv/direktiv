package core

import (
	"crypto/sha256"
	"fmt"
	"io"
	"sync"
)

// nolint:tagliatelle
type ServiceConfig struct {
	Typ       string `json:"type"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`

	FilePath string `json:"filePath"`

	Image string `json:"image"`
	CMD   string `json:"cmd"`
	Size  string `json:"size"`
	Scale int    `json:"scale"`

	Error *string `json:"error"`
}

func (c *ServiceConfig) GetID() string {
	str := fmt.Sprintf("%s-%s-%s-%s", c.Namespace, c.Name, c.Typ, c.FilePath)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("obj%xobj", sh[:10])
}

func (c *ServiceConfig) GetValueHash() string {
	str := fmt.Sprintf("%s-%s-%s-%d", c.Image, c.CMD, c.Size, c.Scale)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("%x", sh[:10])
}

func (c *ServiceConfig) SetDefaults() {
	if c.Size == "" {
		c.Size = "medium"
	}
}

type ServiceStatus struct {
	ServiceConfig
	ID         string `json:"id"`
	Conditions any    `json:"conditions"`
}

type ServiceManager interface {
	Start(done <-chan struct{}, wg *sync.WaitGroup)
	SetServices(list []*ServiceConfig)
	GeAll(namespace string) ([]*ServiceStatus, error)
	GetPods(namespace string, serviceID string) (any, error)
	StreamLogs(namespace string, serviceID string, podID string) (io.ReadCloser, error)
	Kill(namespace string, serviceID string) error
}
