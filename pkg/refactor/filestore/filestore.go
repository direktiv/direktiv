package filestore

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

// Package 'filestore' implements a filesystem that is responsible to store user's projects and files. For each
// direktiv namespace, a 'filestore.Root' should be created to host namespace files and directories.
// 'Root' interface provide all the methods needed to create direktiv namespace files and directories.
// Via 'filestore.Manager' the caller manages the roots, and 'filestore.Root' the caller manages files and directories.

var (
	ErrFileIsNotDirectory = errors.New("ErrFileIsNotDirectory")
	ErrNotFound           = errors.New("ErrNotFound")
)

type Timestamps interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

type Filestore interface {
	CreateRoot(id uuid.UUID) (Root, error)

	GetRoot(id uuid.UUID) (Root, error)
	GetAllRoots() ([]Root, error)

	ForRoot(root Root) RootQuery
	ForFile(file File) FileQuery
	ForRevision(revision Revision) RevisionQuery
	WithContext(ctx context.Context) Filestore
}

type GetFileOpts struct {
	EagerLoad bool
}

type Root interface {
	GetID() uuid.UUID

	Timestamps
}

type RootQuery interface {
	GetFile(path string, opts *GetFileOpts) (File, error)
	CreateFile(path string, typ FileType, dataReader io.Reader) (File, error)
	ReadDirectory(path string) ([]File, error)
	Delete() error
}
