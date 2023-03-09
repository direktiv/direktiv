package filestore

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// Package 'filestore' implements a filesystem that is responsible to store user's projects and files. For each
// direktiv namespace, a 'filestore.Root' should be created to host namespace files and directories.
// 'Root' interface provide all the methods needed to create direktiv namespace files and directories.
// Via 'filestore.Manager' the caller manages the roots, and 'filestore.Root' the caller manages files and directories.

type FileType string

const (
	FileTypeDirectory FileType = "directory"
	FileTypeWorkflow  FileType = "workflow"
	FileTypeFile      FileType = "file"
)

type File interface {
	GetID() uuid.UUID
	GetPath() string
	GetName() string
	GetType() FileType

	GetData(ctx context.Context) (io.ReadCloser, error)

	Delete(ctx context.Context, force bool) error
}

type Filestore interface {
	CreateRoot(ctx context.Context, id uuid.UUID) (Root, error)

	GetRoot(ctx context.Context, id uuid.UUID) (Root, error)
	GetAllRoots(ctx context.Context) ([]Root, error)
}

type GetFileOpts struct {
	EagerLoad bool
}

type Root interface {
	GetID() uuid.UUID
	GetFile(ctx context.Context, path string, opts *GetFileOpts) (File, error)

	CreateFile(ctx context.Context, path string, typ FileType, dataReader io.Reader) (File, error)

	ListPath(ctx context.Context, path string) ([]File, error)
	Delete(ctx context.Context) error
}
