package filestore

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type FileType string

const (
	FileTypeDirectory FileType = "directory"
	FileTypeWorkflow  FileType = "workflow"
	FileTypeFile      FileType = "file"
)

type File struct {
	ID    uuid.UUID
	Path  string
	Depth int
	Typ   FileType

	RootID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileQuery interface {
	GetData(ctx context.Context) (io.ReadCloser, error)
	GetCurrentRevision(ctx context.Context) (*Revision, error)
	GetAllRevisions(ctx context.Context) ([]*Revision, error)
	CreateRevision(ctx context.Context, tags RevisionTags) (*Revision, error)
	Delete(ctx context.Context, force bool) error
	GetRevision(ctx context.Context, id uuid.UUID) (*Revision, error)
	GetRevisionByTag(ctx context.Context, tag string) (*Revision, error)
	SetPath(ctx context.Context, path string) error
}
