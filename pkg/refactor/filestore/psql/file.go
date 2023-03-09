package psql

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

	Data []byte

	RootID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	db *gorm.DB
}

var _ filestore.File = &File{} // Ensures File struct conforms to filestore.File interface.

func (f *File) GetID() uuid.UUID {
	return f.ID
}

func (f *File) GetType() filestore.FileType {
	return f.Typ
}

func (f *File) GetData(ctx context.Context) (io.ReadCloser, error) {
	file := &File{ID: f.ID}
	res := f.db.WithContext(ctx).First(file)
	if res.Error != nil {
		return nil, res.Error
	}
	reader := bytes.NewReader(file.Data)
	readCloser := io.NopCloser(reader)

	return readCloser, nil
}

func (f *File) GetPath() string {
	return f.Path
}

func (f *File) GetName() string {
	return filepath.Base(f.Path)
}

func (f *File) Delete(ctx context.Context, force bool) error {
	res := f.db.WithContext(ctx).Delete(&File{}, f.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}
