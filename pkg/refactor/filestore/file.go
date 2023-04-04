package filestore

import (
	"context"
	"io"
	"path/filepath"
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

func (file *File) Name() string {
	return filepath.Base(file.Path)
}

func (file *File) Dir() string {
	return filepath.Dir(file.Path)
}

type FileQuery interface {
	GetData(ctx context.Context) (io.ReadCloser, error)
	GetCurrentRevision(ctx context.Context) (*Revision, error)
	GetAllRevisions(ctx context.Context) ([]*Revision, error)
	CreateRevision(ctx context.Context, tags RevisionTags, dataReader io.Reader) (*Revision, error)
	Delete(ctx context.Context, force bool) error
	GetRevision(ctx context.Context, id uuid.UUID) (*Revision, error)
	GetRevisionByTag(ctx context.Context, tag string) (*Revision, error)
	SetPath(ctx context.Context, path string) error
}
