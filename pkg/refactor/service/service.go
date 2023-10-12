// nolint
package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("ErrNotFound")

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"
)

type Config struct {
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

func (c *Config) getID() string {
	str := fmt.Sprintf("%s-%s-%s-%s", c.Namespace, c.Name, c.Typ, c.FilePath)
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
	ID string `json:"id"`
	Config
	Conditions any `json:"conditions"`
}
