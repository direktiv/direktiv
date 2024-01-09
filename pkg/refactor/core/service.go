package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"sync"
)

const (
	ServiceTypeNamespace = "namespace-service"
	ServiceTypeWorkflow  = "workflow-service"
)

type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ServicePatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// nolint:tagliatelle
type ServiceConfig struct {
	// identification fields:
	Typ       string `json:"type"`
	Namespace string `json:"namespace"`
	FilePath  string `json:"filePath"`
	Name      string `json:"name"`

	// settings fields:
	Image         string                `json:"image"`
	CMD           string                `json:"cmd"`
	Size          string                `json:"size"`
	Scale         int                   `json:"scale"`
	Envs          []EnvironmentVariable `json:"envs"`
	PostStartExec []string              `json:"post_start_exec"`
	Patches       []ServicePatch        `json:"patches"`

	Error *string `json:"error"`
}

// GetID calculates a unique id string based on identification fields. This id helps in comparison different
// lists of objects.
func (c *ServiceConfig) GetID() string {
	str := fmt.Sprintf("%s-%s-%s", c.Namespace, c.Name, c.FilePath)
	sh := sha256.Sum256([]byte(str + c.Typ))

	whitelist := regexp.MustCompile("[^a-zA-Z0-9]+")
	str = whitelist.ReplaceAllString(str, "-")

	// Prevent too long ids
	// nolint:gomnd
	if len(str) > 50 {
		str = str[:50]
	}

	return fmt.Sprintf("%s-%x", str, sh[:5])
}

// GetValueHash calculates a unique hash string based on the settings fields. This hash helps in comparing
// different lists of objects.
func (c *ServiceConfig) GetValueHash() string {
	str := fmt.Sprintf("%s-%s-%s-%d", c.Image, c.CMD, c.Size, c.Scale)
	for _, v := range c.Envs {
		str += "-" + v.Name + "-" + v.Value
	}
	for _, v := range c.PostStartExec {
		str += "-" + v
	}
	for _, v := range c.Patches {
		str += "-" + v.Op + "-" + v.Path + "-" + fmt.Sprintf("%v", v.Value)
	}

	sh := sha256.Sum256([]byte(str))

	return hex.EncodeToString(sh[:10])
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
	Rebuild(namespace string, serviceID string) error
}
