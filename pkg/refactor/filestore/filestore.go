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

type FileStore interface {
	CreateRoot(ctx context.Context, id uuid.UUID) (*Root, error)

	GetRoot(ctx context.Context, id uuid.UUID) (*Root, error)
	GetAllRoots(ctx context.Context) ([]*Root, error)

	ForRoot(root *Root) RootQuery
	ForFile(file *File) FileQuery
	ForRevision(revision *Revision) RevisionQuery
}

type GetFileOpts struct {
	EagerLoad bool
}

type Root struct {
	ID uuid.UUID

	Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RootQuery interface {
	GetFile(ctx context.Context, path string, opts *GetFileOpts) (*File, error)
	CreateFile(ctx context.Context, path string, typ FileType, dataReader io.Reader) (*File, error)
	ReadDirectory(ctx context.Context, path string) ([]*File, error)
	Delete(ctx context.Context) error
}
