package database

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

type Transaction interface {
	Commit() error
	Rollback() error
}

type Database interface {
	AddTxToCtx(ctx context.Context, tx Transaction) context.Context
	Tx(ctx context.Context) (context.Context, Transaction, error)
	Close() error

	Namespace(ctx context.Context, id uuid.UUID) (*Namespace, error)
	NamespaceByName(ctx context.Context, namespace string) (*Namespace, error)
	Instance(ctx context.Context, id uuid.UUID) (*Instance, error)
	InstanceRuntime(ctx context.Context, id uuid.UUID) (*InstanceRuntime, error)
	NamespaceAnnotation(ctx context.Context, nsID uuid.UUID, key string) (*Annotation, error)
}

type CacheData struct {
	Namespace *Namespace
	// Inodes    []*Inode
	// Workflow  *Workflow
	Ref      *Ref
	Revision *filestore.Revision
	Instance *Instance
	File     *filestore.File
}

func (cached *CacheData) Dir() string {
	return filepath.Dir(cached.File.Path)
}

type HasAttributes interface {
	GetAttributes() map[string]string
}

func GetAttributes(recipientType recipient.RecipientType, a ...HasAttributes) map[string]string {
	m := make(map[string]string)
	m["recipientType"] = string(recipientType)
	for _, x := range a {
		y := x.GetAttributes()
		for k, v := range y {
			m[k] = v
		}
	}
	return m
}

func (cached *CacheData) GetAttributes(recipientType recipient.RecipientType) map[string]string {
	tags := make(map[string]string)
	tags["recipientType"] = string(recipientType)
	if cached.Instance != nil {
		tags["instance-id"] = cached.Instance.ID.String()
		tags["invoker"] = cached.Instance.Invoker
		tags["callpath"] = cached.Instance.CallPath
		tags["workflow"] = GetWorkflow(cached.Instance.As)
	}

	if cached.File != nil {
		tags["workflow-id"] = cached.File.ID.String()
	}

	if cached.Namespace != nil {
		tags["namespace"] = cached.Namespace.Name
		tags["namespace-id"] = cached.Namespace.ID.String()
	}
	return tags
}

func (cached *CacheData) GetAttributesMirror(m *Mirror) map[string]string {
	tags := cached.GetAttributes(recipient.Namespace)
	tags["mirror-id"] = m.ID.String()
	tags["recipientType"] = "mirror"
	return tags
}

func (cached *CacheData) SentLogs(m *Mirror) map[string]string {
	tags := cached.GetAttributes(recipient.Namespace)
	tags["mirror-id"] = m.ID.String()
	tags["recipientType"] = "mirror"
	return tags
}

func GetWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
