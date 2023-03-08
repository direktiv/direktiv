package filesystem

import "context"

// Package 'filesystem' implements a filesystem that is responsible to store user's projects and files.

type File interface {
	GetPath() string
	GetPayload() []byte
	GetName() string
	GetIsDirectory() bool
	Delete(ctx context.Context, forceDelete bool) error
}

type Filesystem interface {
	CreateNamespace(ctx context.Context, namespace string) (Namespace, error)
	GetNamespace(ctx context.Context, namespace string) (Namespace, error)
	DeleteNamespace(ctx context.Context, namespace string) error
	GetAllNamespaces(ctx context.Context) ([]Namespace, error)
}

type Namespace interface {
	CreateFile(ctx context.Context, path string, typ string, payload []byte) (File, error)
	GetFile(ctx context.Context, path string) (File, error)
	ListPath(ctx context.Context, path string) ([]File, error)
	GetName() string
}
