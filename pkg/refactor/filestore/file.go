package filestore

import (
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
	GetData() (io.ReadCloser, error)
	GetCurrentRevision() (Revision, error)
	CreateRevision(tags RevisionTags) (Revision, error)
	Delete(force bool) error
}
