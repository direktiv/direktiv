package database

import (
	"context"
	"path/filepath"
	"strings"

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
