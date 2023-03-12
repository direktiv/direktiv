package filestore

import (
	"context"
	"io"

	"github.com/google/uuid"
)

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

	Timestamps
}

type FileQuery interface {
	GetData(ctx context.Context) (io.ReadCloser, error)
	GetCurrentRevision(ctx context.Context) (Revision, error)
	CreateRevision(ctx context.Context, tags RevisionTags) (Revision, error)
	Delete(ctx context.Context, force bool) error
}
