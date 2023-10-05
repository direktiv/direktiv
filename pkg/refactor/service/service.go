// nolint
package service

import (
	"crypto/sha256"
	"fmt"
)

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"
)

type Config struct {
	Namespace string  `json:"namespace"`
	Name      *string `json:"name"`

	ServicePath  *string `json:"servicePath"`
	WorkflowPath *string `json:"workflowPath"`

	Image string `json:"image"`
	CMD   string `json:"cmd"`
	Size  string `json:"size"`
	Scale int    `json:"scale"`

	Error *string `json:"error"`
}

func (c *Config) getID() string {
	str := fmt.Sprintf("%s-%v-%v-%v", c.Namespace, c.Name, c.ServicePath, c.WorkflowPath)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("obj%xobj", sh[:10])
}

func (c *Config) getValueHash() string {
	str := fmt.Sprintf("%s-%s-%s-%d", c.Image, c.CMD, c.Size, c.Scale)
	sh := sha256.Sum256([]byte(str))

	return fmt.Sprintf("%x", sh[:10])
}

type Status interface {
	getConditions() any
	getID() string
	getValueHash() string
}

type ConfigStatus struct {
	ID         string  `json:"id"`
	Config     *Config `json:"config"`
	Conditions any     `json:"conditions"`
}
