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
	rootID       uuid.UUID
	checksumFunc filestore.CalculateChecksumFunc
	db           *gorm.DB
}

func (q *RootQuery) IsEmpty(ctx context.Context) (bool, error) {
	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return false, err
	}
	count := 0
	tx := q.db.Raw("SELECT count(id) FROM files WHERE root_id = ?", q.rootID).Scan(&count)
	if tx.Error != nil {
		return false, tx.Error
	}

	return count == 0, nil
}

var _ filestore.RootQuery = &RootQuery{} // Ensures RootQuery struct conforms to filestore.RootQuery interface.

func (q *RootQuery) Delete(ctx context.Context) error {
	res := q.db.WithContext(ctx).Delete(&filestore.Root{ID: q.rootID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

//nolint:ireturn
func (q *RootQuery) CreateFile(ctx context.Context, path string, typ filestore.FileType, dataReader io.Reader) (*filestore.File, *filestore.Revision, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, nil, filestore.ErrInvalidPathParameter
	}

	// check if root exists.
	if err = q.checkRootExists(ctx); err != nil {
		return nil, nil, err
	}

	count := 0
	tx := q.db.Raw("SELECT count(id) FROM files WHERE root_id = ? AND path = ?", q.rootID, path).Scan(&count)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	if count > 0 {
		return nil, nil, filestore.ErrPathAlreadyExists
	}

	parentDir := filepath.Dir(path)
	if parentDir != "/" {
		count = 0
		tx = q.db.Raw("SELECT count(id) FROM files WHERE root_id = ? AND typ = ? AND path = ?", q.rootID, filestore.FileTypeDirectory, parentDir).Scan(&count)
		if tx.Error != nil {
			return nil, nil, tx.Error
		}
		if count == 0 {
			return nil, nil, filestore.ErrNoParentDirectory
		}
	}

	// first, we need to create a file entry for this new file.
	f := &filestore.File{
		ID:     uuid.New(),
		Path:   path,
		Depth:  filestore.ParseDepth(path),
		Typ:    typ,
		RootID: q.rootID,
	}

	res := q.db.WithContext(ctx).Create(f)
	if res.Error != nil {
		return nil, nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	if typ == filestore.FileTypeDirectory {
		return f, nil, nil
	}

	// second, now we need to create a revision entry for this new file.
	var data []byte
	data, err = io.ReadAll(dataReader)
	if err != nil {
		return nil, nil, fmt.Errorf("create io error, %w", err)
	}
	rev := &filestore.Revision{
		ID:   uuid.New(),
		Tags: "",

		FileID:    f.ID,
		IsCurrent: true,

		Data:     data,
		Checksum: string(q.checksumFunc(data)),
	}
	res = q.db.WithContext(ctx).Create(rev)
	if res.Error != nil {
		return nil, nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return f, rev, nil
}

//nolint:ireturn
func (q *RootQuery) GetFile(ctx context.Context, path string) (*filestore.File, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, filestore.ErrInvalidPathParameter
	}

	// check if root exists.
	if err = q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	f := &filestore.File{}
	path = filepath.Clean(path)

	res := q.db.WithContext(ctx).Where("root_id", q.rootID).Where("path = ?", path).First(f)
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
		return nil, filestore.ErrInvalidPathParameter
	}

	// check if root exists.
	if err = q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	if path == "" {
		path = "/"
	}

	// check if path is a directory and exists.
	if path != "/" {
		count := 0
		tx := q.db.Raw("SELECT count(id) FROM files WHERE root_id = ? AND typ = ? AND path = ?", q.rootID, filestore.FileTypeDirectory, path).Scan(&count)
		if tx.Error != nil {
			return nil, err
		}
		if count != 1 {
			return nil, filestore.ErrPathIsNotDirectory
		}
	}

	res := q.db.WithContext(ctx).
		// Don't include file 'data' in the query. File data can be retrieved with file.GetData().
		Select("id", "path", "depth", "typ", "root_id", "created_at", "updated_at").
		Where("root_id", q.rootID).
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

//nolint:ireturn
func (q *RootQuery) CalculateChecksumsMap(ctx context.Context, path string) (map[string]string, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, filestore.ErrInvalidPathParameter
	}

	// check if root exists.
	if err = q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	type Result struct {
		Path     string
		Typ      string
		Checksum string
	}

	var resultList []Result

	res := q.db.WithContext(ctx).
		// Don't include file 'data' in the query. File data can be retrieved with file.GetData().
		Raw(`SELECT f.path, f.typ, r.checksum 
				 FROM files AS f 
				     LEFT JOIN revisions r 
				         ON r.file_id = f.id AND r.is_current = true 
				 WHERE f.depth = ? AND path LIKE ?`,
			filestore.ParseDepth(path)+1, addTrailingSlash(path)+"%").Scan(&resultList)

	if res.Error != nil {
		return nil, res.Error
	}

	result := make(map[string]string)

	for _, item := range resultList {
		result[item.Path] = item.Checksum
	}

	return result, nil
}

func (q *RootQuery) checkRootExists(ctx context.Context) error {
	n := &filestore.Root{ID: q.rootID}
	res := q.db.WithContext(ctx).First(n)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
