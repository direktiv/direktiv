package varstore

import (
	"context"
	"io"
)

type VarStorage interface {
	Store(ctx context.Context, key string, scope ...string) (io.WriteCloser, error)
	Retrieve(ctx context.Context, key string, scope ...string) (VarReader, error)
	List(ctx context.Context, scope ...string) ([]VarInfo, error)
	Delete(ctx context.Context, key string, scope ...string) error
	DeleteAllInScope(ctx context.Context, scope ...string) error
	io.Closer
}

type VarReader interface {
	io.Reader
	io.Closer
	Size() int64
}

type VarInfo interface {
	Key() string
	Size() int64
}
