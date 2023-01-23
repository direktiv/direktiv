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
	Tx(ctx context.Context) (Transaction, error)
	Close() error

	Namespace(ctx context.Context, tx Transaction, id uuid.UUID) (*Namespace, error)
	NamespaceByName(ctx context.Context, tx Transaction, namespace string) (*Namespace, error)
	Inode(ctx context.Context, tx Transaction, id uuid.UUID) (*Inode, error)
	Workflow(ctx context.Context, tx Transaction, id uuid.UUID) (*Workflow, error)
	Revision(ctx context.Context, tx Transaction, id uuid.UUID) (*Revision, error)
	Instance(ctx context.Context, tx Transaction, id uuid.UUID) (*Instance, error)
	InstanceRuntime(ctx context.Context, tx Transaction, id uuid.UUID) (*InstanceRuntime, error)
	NamespaceAnnotation(ctx context.Context, tx Transaction, nsID uuid.UUID, key string) (*Annotation, error)
	InodeAnnotation(ctx context.Context, tx Transaction, inodeID uuid.UUID, key string) (*Annotation, error)
	WorkflowAnnotation(ctx context.Context, tx Transaction, wfID uuid.UUID, key string) (*Annotation, error)
	InstanceAnnotation(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*Annotation, error)
	ThreadVariables(ctx context.Context, tx Transaction, instID uuid.UUID) ([]*VarRef, error)
	NamespaceVariableRef(ctx context.Context, tx Transaction, nsID uuid.UUID, key string) (*VarRef, error)
	WorkflowVariableRef(ctx context.Context, tx Transaction, wfID uuid.UUID, key string) (*VarRef, error)
	InstanceVariableRef(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*VarRef, error)
	ThreadVariableRef(ctx context.Context, tx Transaction, instID uuid.UUID, key string) (*VarRef, error)
	VariableData(ctx context.Context, tx Transaction, id uuid.UUID, load bool) (*VarData, error)
	Mirror(ctx context.Context, tx Transaction, id uuid.UUID) (*Mirror, error)
}

type CacheData struct {
	Namespace *Namespace
	Inodes    []*Inode
	Workflow  *Workflow
	Ref       *Ref
	Revision  *Revision
	Instance  *Instance
}

func (cached *CacheData) Path() string {
	var elems []string
	for _, ino := range cached.Inodes {
		elems = append(elems, ino.Name)
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
