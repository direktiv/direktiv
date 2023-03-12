package psql

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func addTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}

	return path
}

type RootQuery struct {
	root filestore.Root
	db   *gorm.DB
	ctx  context.Context
}

var _ filestore.RootQuery = &RootQuery{} // Ensures RootQuery struct conforms to filestore.RootQuery interface.

func (q *RootQuery) Delete() error {
	res := q.db.WithContext(q.ctx).Delete(&Root{ID: q.root.GetID()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

//nolint:ireturn
func (q *RootQuery) CreateFile(path string, typ filestore.FileType, dataReader io.Reader) (filestore.File, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, fmt.Errorf("create validation error, %w", err)
	}

	// first, we need to create a file entry for this new file.
	f := &File{
		ID:     uuid.New(),
		Path:   path,
		Depth:  filestore.ParseDepth(path),
		Typ:    typ,
		RootID: q.root.GetID(),
		db:     q.db,
	}

	res := q.db.WithContext(q.ctx).Create(f)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	if typ == filestore.FileTypeDirectory {
		return f, nil
	}

	// second, now we need to create a revision entry for this new file.
	var data []byte
	if dataReader != nil {
		data, err = io.ReadAll(dataReader)
		if err != nil {
			return nil, fmt.Errorf("create io error, %w", err)
		}
	}

	rev := &Revision{
		ID:   uuid.New(),
		Tags: "",

		FileID:    f.ID,
		IsCurrent: true,

		Data: data,
	}
	res = q.db.WithContext(q.ctx).Create(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return f, nil
}

//nolint:ireturn
func (q *RootQuery) GetFile(path string, opts *filestore.GetFileOpts) (filestore.File, error) {
	f := &File{}
	path = filepath.Clean(path)

	res := q.db.WithContext(q.ctx).Where("root_id", q.root.GetID()).Where("path = ?", path).First(f)
	if res.Error != nil {
		return nil, res.Error
	}
	f.db = q.db

	return f, nil
}

//nolint:ireturn
func (q *RootQuery) ListPath(path string) ([]filestore.File, error) {
	var list []File
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, fmt.Errorf("create error, %w", err)
	}

	res := q.db.WithContext(q.ctx).
		// Don't include file 'data' in the query. File data can be retrieved with file.GetData().
		Select("id", "path", "depth", "root_id", "created_at", "updated_at").
		Where("root_id", q.root.GetID()).
		Where("depth", filestore.ParseDepth(path)+1).
		Where("path LIKE ?", addTrailingSlash(path)+"%"). // trailing slash necessary otherwise "/a" will receive children for both "/a" and "/abc".
		Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}
	var files []filestore.File
	for i := range list {
		list[i].db = q.db
		files = append(files, &list[i])
	}

	return files, nil
}
