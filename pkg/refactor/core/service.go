package core

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"sync"
)

const (
	ServiceTypeNamespace = "namespace-service"
	ServiceTypeWorkflow  = "workflow-service"
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
	var str, prefix string
	if c.Typ == ServiceTypeNamespace {
		str = fmt.Sprintf("%s:%s", c.Namespace, c.Name)
	} else {
		path := strings.Trim(c.FilePath, "/")
		path = strings.TrimSuffix(path, ".yaml")
		path = strings.TrimSuffix(path, ".yml")
		str = fmt.Sprintf("%s/%s:%s", c.Namespace, path, c.Name)
	}

	sh := sha256.Sum256([]byte(str))

	// NOTES:
	// 		Only the hash really matters. The prefix is just for human readability.
	//		Restrictions are usually related to DNS subdomain naming.
	prefix = str
	prefix = strings.SplitN(prefix, ":", 2)[0] //nolint:gomnd // NOTE: excluding the name because we're currently not strict about naming services and it will be a pain to sanitize.
	prefix = strings.ReplaceAll(prefix, "/", "-")
	prefix = strings.ReplaceAll(prefix, "_", "-")
	prefix = strings.ReplaceAll(prefix, ".", "-")
	prefix = strings.ToLower(prefix)

	// Has a maximum length of 63. But I can't remember if knative wants to use some of it, so I'm using less of the available limit to be safe.
	maxLen := 50 //nolint:gomnd
	if len(prefix) > maxLen {
		prefix = prefix[:maxLen]
	}

	return fmt.Sprintf("%s-%x", prefix, sh[:10])
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
