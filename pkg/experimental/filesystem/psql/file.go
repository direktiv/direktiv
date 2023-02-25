package psql

import (
	"context"
	"fmt"
	"github.com/direktiv/direktiv/pkg/experimental/filesystem"
	"gorm.io/gorm"
	"path/filepath"
	"time"
)

type File struct {
	ID          int64
	Path        string
	Depth       int
	Payload     []byte
	IsDirectory bool

	NamespaceID int64

	CreatedAt time.Time
	UpdatedAt time.Time

	db *gorm.DB
}

func (f File) GetPath() string {
	return f.Path
}

func (f File) GetPayload() []byte {
	return f.Payload
}

func (f File) GetName() string {
	return filepath.Base(f.Path)
}

func (f File) GetIsDirectory() bool {
	return f.IsDirectory
}

func (f File) Delete(ctx context.Context, forceDelete bool) error {
	res := f.db.WithContext(ctx).Delete(&File{}, f.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}
	return nil
}

var _ filesystem.File = &File{}
