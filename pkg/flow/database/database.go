package database

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
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
	ThreadVariables(ctx context.Context, instID uuid.UUID) ([]*VarRef, error)
	NamespaceVariableRef(ctx context.Context, nsID uuid.UUID, key string) (*VarRef, error)
	WorkflowVariableRef(ctx context.Context, wfID uuid.UUID, key string) (*VarRef, error)
	InstanceVariableRef(ctx context.Context, instID uuid.UUID, key string) (*VarRef, error)
	ThreadVariableRef(ctx context.Context, instID uuid.UUID, key string) (*VarRef, error)
	VariableData(ctx context.Context, id uuid.UUID, load bool) (*VarData, error)
}

type CacheData struct {
	Namespace *Namespace
	Inodes    []*Inode
	Workflow  *Workflow
	Ref       *Ref
	Revision  *Revision
	Instance  *Instance
}

func (cached *CacheData) Parent() *CacheData {
	return &CacheData{
		Namespace: cached.Namespace,
		Inodes:    cached.Inodes[:len(cached.Inodes)-1],
		Workflow:  cached.Workflow,
		Ref:       cached.Ref,
		Revision:  cached.Revision,
		Instance:  cached.Instance,
	}
}

func (cached *CacheData) Path() string {
	var elems []string
	for _, ino := range cached.Inodes {
		elems = append(elems, ino.Name)
	}

	if len(elems) == 1 {
		return "/"
	}

	return strings.Join(elems, "/")
}

func (cached *CacheData) Dir() string {
	return filepath.Dir(cached.Path())
}

func (cached *CacheData) Reset() {
	cached.Inodes = make([]*Inode, 0)
}

func (cached *CacheData) Inode() *Inode {
	return cached.Inodes[len(cached.Inodes)-1]
}

func (cached *CacheData) ParentInode() *Inode {
	return cached.Inodes[len(cached.Inodes)-2]
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

	if cached.Workflow != nil {
		tags["workflow-id"] = cached.Workflow.ID.String()
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
