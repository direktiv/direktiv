// nolint
package service

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
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
	str := fmt.Sprintf("%s-%s-%s-%s", c.Namespace, c.Name, c.ServicePath, c.WorkflowPath)
	v, err := hashstructure.Hash(str, hashstructure.FormatV2, nil)
	if err != nil {
		panic("unexpected hashstructure.Hash error: " + err.Error())
	}

	return fmt.Sprintf("obj-%d-obj", v)
}

func (c *Config) getValueHash() string {
	str := fmt.Sprintf("%s-%s-%s-%d", c.Image, c.CMD, c.Size, c.Scale)
	v, err := hashstructure.Hash(str, hashstructure.FormatV2, nil)
	if err != nil {
		panic("unexpected hashstructure.Hash error: " + err.Error())
	}

	return strconv.Itoa(int(v))
}

type Status interface {
	getConditions() any
	getID() string
	getValueHash() string
}

type ConfigStatus struct {
	Config     *Config `json:"config"`
	Conditions any     `json:"conditions,omitempty"`
}
