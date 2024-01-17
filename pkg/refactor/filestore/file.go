package filestore

import (
	"context"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// FileType represents file types. Filestore files are basically two types, ordinary files and directories.
type FileType string

const (
	// FileTypeWorkflow is special file type as we handle workflow differently.
	FileTypeWorkflow  FileType = "workflow"
	FileTypeEndpoint  FileType = "endpoint"
	FileTypeConsumer  FileType = "consumer"
	FileTypeService   FileType = "service"
	FileTypeFile      FileType = "file"
	FileTypeDirectory FileType = "directory"
)

const (
	Latest = "latest"
)

// File represents a file in the filestore, File can be either ordinary file or directory.
type File struct {
	ID uuid.UUID
	// Path is the full path of the file, files and directories are only different when they have different paths. As
	// in typical filesystems, paths are unique within the filesystem.
	Path string

	// Depth tells how many levels deep the file in the filesystem. This field is needed for sql querying purposes.
	Depth int
	Typ   FileType

	// Root is a filestore instance, users can create multiple filestore roots and RootID tells which root the file
	// belongs too.
	RootID uuid.UUID

	MIMEType string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Name gets file base name.
func (file *File) Name() string {
	return filepath.Base(file.Path)
}

// Dir gets file directory.
func (file *File) Dir() string {
	return filepath.Dir(file.Path)
}

// FileQuery performs different queries associated to a file.
type FileQuery interface {
	// GetData returns reader for the file, this method is not applicable for directory file type.
	GetData(ctx context.Context) ([]byte, error)

	// GetCurrentRevision returns current file revision, this method is not applicable for directory file type.
	GetCurrentRevision(ctx context.Context) (*Revision, error)

	// GetAllRevisions lists all file revisions, this method is not applicable for directory file type.
	GetAllRevisions(ctx context.Context) ([]*Revision, error)

	// CreateRevision creates a new file revision, this method is not applicable for directory file type.
	CreateRevision(ctx context.Context, data []byte) (*Revision, error)

	// Delete deletes the file (or the directory).
	Delete(ctx context.Context, force bool) error

	// GetRevision queries a file revision by either:
	// 		reference = 'latest' -> this will return the current revision,
	// or:	reference = any tag string -> this will return a tagged revision,
	// or:	reference = any uuid string -> this will return a revision by id.
	// This method is not applicable for directory file type.
	GetRevision(ctx context.Context) (*Revision, error)

	// SetPath sets a new path for the file, this method is used to rename files and directories or move them
	// to a new location. Param path should be a new path that doesn't already exist and the directory of Param path
	// should already exist.
	SetPath(ctx context.Context, path string) error
}
