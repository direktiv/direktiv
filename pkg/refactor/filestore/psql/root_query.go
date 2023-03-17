package psql

import (
	"context"
	"crypto/sha256"
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
	root *filestore.Root
	db   *gorm.DB
}

var _ filestore.RootQuery = &RootQuery{} // Ensures RootQuery struct conforms to filestore.RootQuery interface.

func (q *RootQuery) Delete(ctx context.Context) error {
	res := q.db.WithContext(ctx).Delete(&filestore.Root{ID: q.root.ID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

//nolint:ireturn
func (q *RootQuery) CreateFile(ctx context.Context, path string, typ filestore.FileType, dataReader io.Reader) (*filestore.File, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, fmt.Errorf("create validation error, %w", err)
	}

	// first, we need to create a file entry for this new file.
	f := &filestore.File{
		ID:     uuid.New(),
		Path:   path,
		Depth:  filestore.ParseDepth(path),
		Typ:    typ,
		RootID: q.root.ID,
	}

	res := q.db.WithContext(ctx).Create(f)
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
	var checksum [32]byte
	data, err = io.ReadAll(dataReader)
	if err != nil {
		return nil, fmt.Errorf("create io error, %w", err)
	}
	checksum = sha256.Sum256(data)

	rev := &filestore.Revision{
		ID:   uuid.New(),
		Tags: "",

		FileID:    f.ID,
		IsCurrent: true,

		Data:     data,
		Checksum: string(checksum[:]),
	}
	res = q.db.WithContext(ctx).Create(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return f, nil
}

//nolint:ireturn
func (q *RootQuery) GetFile(ctx context.Context, path string) (*filestore.File, error) {
	f := &filestore.File{}
	path = filepath.Clean(path)

	res := q.db.WithContext(ctx).Where("root_id", q.root.ID).Where("path = ?", path).First(f)
	if res.Error != nil {
		return nil, res.Error
	}

	return f, nil
}

//nolint:ireturn
func (q *RootQuery) ReadDirectory(ctx context.Context, path string) ([]*filestore.File, error) {
	var list []filestore.File
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, fmt.Errorf("create error, %w", err)
	}

	res := q.db.WithContext(ctx).
		// Don't include file 'data' in the query. File data can be retrieved with file.GetData().
		Select("id", "path", "depth", "root_id", "created_at", "updated_at").
		Where("root_id", q.root.ID).
		Where("depth", filestore.ParseDepth(path)+1).
		Where("path LIKE ?", addTrailingSlash(path)+"%"). // trailing slash necessary otherwise "/a" will receive children for both "/a" and "/abc".
		Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}
	var files []*filestore.File
	for i := range list {
		files = append(files, &list[i])
	}

	return files, nil
}
