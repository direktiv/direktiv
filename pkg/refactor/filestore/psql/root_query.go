package psql

import (
	"context"
	"errors"
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

func (q *RootQuery) CropFilesAndDirectories(ctx context.Context, excludePaths []string) error {
	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return err
	}

	var allPaths []string

	res := q.db.WithContext(ctx).Table("filesystem_files").Select("path").Where("root_id", q.rootID).Find(&allPaths)
	if res.Error != nil {
		return res.Error
	}

	pathsToRemove := []string{}
	for _, pathInRoot := range allPaths {
		remove := true
		for _, excludePath := range excludePaths {
			if strings.HasPrefix(excludePath, pathInRoot) {
				remove = false

				break
			}
			if excludePath == pathInRoot {
				remove = false

				break
			}
		}
		if remove {
			pathsToRemove = append(pathsToRemove, pathInRoot)
		}
	}

	res = q.db.WithContext(ctx).Exec("DELETE FROM filesystem_files WHERE root_id = ? AND path IN ?", q.rootID, pathsToRemove)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != int64(len(pathsToRemove)) {
		return fmt.Errorf("unexpedted delete from filesystem_files count, got: %d, want: %d", res.RowsAffected, len(pathsToRemove))
	}

	return nil
}

func (q *RootQuery) ListAllFiles(ctx context.Context) ([]*filestore.File, error) {
	var list []*filestore.File

	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	res := q.db.WithContext(ctx).Table("filesystem_files").Where("root_id", q.rootID).Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

func (q *RootQuery) IsEmptyDirectory(ctx context.Context, path string) (bool, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return false, filestore.ErrInvalidPathParameter
	}

	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return false, err
	}

	// check if dir exists.
	count := 0
	tx := q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ? AND typ = ? AND path = ?",
		q.rootID, filestore.FileTypeDirectory, path).
		Scan(&count)

	if tx.Error != nil {
		return false, tx.Error
	}
	if count == 0 {
		return false, filestore.ErrNotFound
	}

	// check if there are child entries.
	if path == "/" {
		count = 0
		tx = q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ?", q.rootID).Scan(&count)
		if tx.Error != nil {
			return false, tx.Error
		}

		return count == 1, nil
	}

	// check if there are child entries.
	count = 0
	tx = q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ? AND path LIKE ?", q.rootID, path+"/%").Scan(&count)
	if tx.Error != nil {
		return false, tx.Error
	}

	return count == 0, nil
}

var _ filestore.RootQuery = &RootQuery{} // Ensures RootQuery struct conforms to filestore.RootQuery interface.

func (q *RootQuery) Delete(ctx context.Context) error {
	res := q.db.WithContext(ctx).Table("filesystem_roots").Delete(&filestore.Root{}, q.rootID)
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
	tx := q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ? AND path = ?", q.rootID, path).Scan(&count)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	if count > 0 {
		return nil, nil, filestore.ErrPathAlreadyExists
	}

	parentDir := filepath.Dir(path)
	if path != "/" {
		count = 0
		tx = q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ? AND typ = ? AND path = ?", q.rootID, filestore.FileTypeDirectory, parentDir).Scan(&count)
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
		Depth:  filestore.GetPathDepth(path),
		Typ:    typ,
		RootID: q.rootID,
	}

	res := q.db.WithContext(ctx).Table("filesystem_files").Create(f)
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
	if dataReader == nil {
		return nil, nil, fmt.Errorf("parameter dataReader is nil with FileTypeFile")
	}
	data, err = io.ReadAll(dataReader)
	if err != nil {
		return nil, nil, fmt.Errorf("reading dataReader, error: %w", err)
	}

	rev := &filestore.Revision{
		ID:   uuid.New(),
		Tags: "",

		FileID:    f.ID,
		IsCurrent: true,

		Data:     data,
		Checksum: string(q.checksumFunc(data)),
	}
	res = q.db.WithContext(ctx).Table("filesystem_revisions").Create(rev)
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

	res := q.db.WithContext(ctx).Table("filesystem_files").Where("root_id", q.rootID).Where("path = ?", path).First(f)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("file '%s': %w", path, filestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return f, nil
}

//nolint:ireturn
func (q *RootQuery) ReadDirectory(ctx context.Context, path string) ([]*filestore.File, error) {
	var list []*filestore.File
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, filestore.ErrInvalidPathParameter
	}

	// check if root exists.
	if err = q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	// check if path is a directory and exists.
	if path != "/" {
		count := 0
		tx := q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ? AND typ = ? AND path = ?", q.rootID, filestore.FileTypeDirectory, path).Scan(&count)
		if tx.Error != nil {
			return nil, err
		}
		if count != 1 {
			return nil, filestore.ErrNotFound
		}
	}

	res := q.db.WithContext(ctx).
		Table("filesystem_files").
		// Don't include file 'data' in the query. File data can be retrieved with file.GetData().
		Select("id", "path", "depth", "typ", "root_id", "created_at", "updated_at").
		Where("root_id", q.rootID).
		Where("depth", filestore.GetPathDepth(path)+1).
		Where("path LIKE ?", addTrailingSlash(path)+"%"). // trailing slash necessary otherwise "/a" will receive children for both "/a" and "/abc".
		Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

//nolint:ireturn
func (q *RootQuery) CalculateChecksumsMap(ctx context.Context) (map[string]string, error) {
	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	type Result struct {
		Path     string
		Checksum string
	}

	var resultList []Result

	res := q.db.WithContext(ctx).
		// Don't include file 'data' in the query. File data can be retrieved with file.GetData().
		Raw(`SELECT f.path, r.checksum 
				 FROM filesystem_files AS f 
				 LEFT JOIN filesystem_revisions r 
					ON r.file_id = f.id AND r.is_current = true
      	     	 WHERE f.root_id = ?
					`, q.rootID).Scan(&resultList)

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
	n := &filestore.Root{}
	res := q.db.WithContext(ctx).Table("filesystem_roots").Where("id", q.rootID).First(n)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("root not found, id: '%s', err: %w", q.rootID, filestore.ErrNotFound)
	}
	if res.Error != nil {
		return res.Error
	}

	return nil
}
