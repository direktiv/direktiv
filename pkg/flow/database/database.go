package database

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
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
	ThreadVariables(ctx context.Context, instID uuid.UUID) ([]*VarRef, error)
	NamespaceVariableRef(ctx context.Context, nsID uuid.UUID, key string) (*VarRef, error)
	WorkflowVariableRef(ctx context.Context, wfID uuid.UUID, key string) (*VarRef, error)
	InstanceVariableRef(ctx context.Context, instID uuid.UUID, key string) (*VarRef, error)
	ThreadVariableRef(ctx context.Context, instID uuid.UUID, key string) (*VarRef, error)
	VariableData(ctx context.Context, id uuid.UUID, load bool) (*VarData, error)
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
	GetAttributes() map[string]interface{}
}

func GetAttributes(recipientType recipient.RecipientType, a ...HasAttributes) map[string]interface{} {
	m := make(map[string]interface{})
	m["sender_type"] = recipientType

	for _, x := range a {
		y := x.GetAttributes()
		for k, v := range y {
			m[k] = v
		}
	}
	return m
}

func (cached *CacheData) GetAttributes(recipientType recipient.RecipientType) map[string]interface{} {
	tags := make(map[string]interface{})
	callpath := ""
	tags["sender_type"] = recipientType
	switch recipientType {
	case "namespace":
		tags["sender"] = cached.Namespace.ID
	case "instance":
		tags["sender"] = cached.Instance.ID
		callpath = internallogger.AppendInstanceID(cached.Instance.CallPath, cached.Instance.ID.String())
		rootInstance, err := internallogger.GetRootinstanceID(callpath)
		if err != nil {
			panic("malformed callpath")
		}
		tags["root_instance_id"] = rootInstance
	case "workflow":
		tags["sender"] = cached.File.ID
	}
	if cached.Instance != nil {
		tags["instance_logs"] = cached.Instance.ID
		tags["invoker"] = cached.Instance.Invoker
		tags["workflow"] = GetWorkflow(cached.Instance.As)
	}
	if callpath != "" {
		tags["log_instance_call_path"] = callpath
	}
	if cached.File != nil {
		tags["workflow_id"] = cached.File.ID
	}

	if cached.Namespace != nil {
		tags["namespace"] = cached.Namespace.Name
		tags["namespace_logs"] = cached.Namespace.ID
	}
	return tags
}

func (cached *CacheData) GetAttributesMirror(m *Mirror) map[string]interface{} {
	tags := cached.GetAttributes("namespace")
	tags["mirror_activity_id"] = m.ID
	tags["sender_type"] = "mirror"
	tags["sender"] = m.ID
	return tags
}

func (cached *CacheData) SentLogs(m *Mirror) map[string]interface{} {
	tags := cached.GetAttributes("namespace")
	tags["mirror_activity_id"] = m.ID
	tags["sender"] = m.ID
	tags["sender_type"] = "mirror"
	return tags
}

func GetWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
