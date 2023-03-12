package psql

import (
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID    uuid.UUID
	Path  string
	Depth int
	Typ   filestore.FileType

	Revisions []Revision `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	RootID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	db *gorm.DB
}

var _ filestore.File = &File{} // Ensures File struct conforms to filestore.File interface.

func (f *File) GetCreatedAt() time.Time {
	return f.CreatedAt
}

func (f *File) GetUpdatedAt() time.Time {
	return f.UpdatedAt
}

func (f *File) GetID() uuid.UUID {
	return f.ID
}

func (f *File) GetType() filestore.FileType {
	return f.Typ
}

func (f *File) GetPath() string {
	return f.Path
}

func (f *File) GetName() string {
	return filepath.Base(f.Path)
}
