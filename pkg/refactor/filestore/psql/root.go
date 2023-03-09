package psql

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Root struct {
	ID uuid.UUID

	Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time

	db *gorm.DB
}

var _ filestore.Root = &Root{} // Ensures Root struct conforms to filestore.Root interface.

type RootList []*Root

func (r *Root) GetID() uuid.UUID {
	return r.ID
}

func (r *Root) Delete(ctx context.Context) error {
	res := r.db.WithContext(ctx).Delete(&Root{ID: r.ID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

//nolint:ireturn
func (r *Root) CreateFile(ctx context.Context, path string, typ filestore.FileType, dataReader io.Reader) (filestore.File, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, fmt.Errorf("create validation error, %w", err)
	}

	var data []byte
	if dataReader != nil {
		data, err = io.ReadAll(dataReader)
		if err != nil {
			return nil, fmt.Errorf("create io error, %w", err)
		}
	}

	f := &File{
		ID:     uuid.New(),
		Path:   path,
		Depth:  filestore.ParseDepth(path),
		Data:   data,
		Typ:    typ,
		RootID: r.ID,
		db:     r.db,
	}
	res := r.db.WithContext(ctx).Create(f)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return f, nil
}

//nolint:ireturn
func (r *Root) GetFile(ctx context.Context, path string, opts *filestore.GetFileOpts) (filestore.File, error) {
	f := &File{}
	path = filepath.Clean(path)

	res := r.db.WithContext(ctx).Where("root_id", r.ID).Where("path = ?", path).First(f)
	if res.Error != nil {
		return nil, res.Error
	}
	f.db = r.db

	return f, nil
}

func addTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

//nolint:ireturn
func (r *Root) ListPath(ctx context.Context, path string) ([]filestore.File, error) {
	var list []File
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, fmt.Errorf("create error, %w", err)
	}

	res := r.db.WithContext(ctx).
		// Don't include file 'data' in the query. File data can be retrieved with file.GetData().
		Select("id", "path", "depth", "root_id", "created_at", "updated_at").
		Where("root_id", r.ID).
		Where("depth", filestore.ParseDepth(path)+1).
		Where("path LIKE ?", addTrailingSlash(path)+"%"). // trailing slash necessary otherwise "/a" will receive children for both "/a" and "/abc".
		Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}
	var files []filestore.File
	for i := range list {
		list[i].db = r.db
		files = append(files, &list[i])
	}

	return files, nil
}
