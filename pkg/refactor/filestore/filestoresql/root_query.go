package filestoresql

import (
	"context"
	"errors"
	"fmt"
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
	root         *filestore.Root
	namespace    string
}

func (q *RootQuery) ListAllFiles(ctx context.Context) ([]*filestore.File, error) {
	var list []*filestore.File

	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	res := q.db.WithContext(ctx).Table("filesystem_files").Where("root_id", q.rootID).Order("path ASC").Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

func (q *RootQuery) ListDirektivFiles(ctx context.Context) ([]*filestore.File, error) {
	var list []*filestore.File

	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return nil, err
	}

	res := q.db.WithContext(ctx).Raw(`
						SELECT * 
						FROM filesystem_files 
						WHERE root_id=? AND typ <> 'directory' AND typ <> 'file'
						`, q.rootID).Find(&list)
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
	// check if root exists.
	if err := q.checkRootExists(ctx); err != nil {
		return err
	}

	res := q.db.WithContext(ctx).Exec(`DELETE FROM filesystem_roots WHERE id = ?`, q.rootID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *RootQuery) CreateFile(ctx context.Context, path string, typ filestore.FileType, mimeType string, data []byte) (*filestore.File, *filestore.Revision, error) {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", filestore.ErrInvalidPathParameter, err)
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
		ID:       uuid.New(),
		Path:     path,
		Depth:    filestore.GetPathDepth(path),
		Typ:      typ,
		RootID:   q.rootID,
		MIMEType: mimeType,
	}

	res := q.db.WithContext(ctx).Table("filesystem_files").Create(f)
	if res.Error != nil {
		return nil, nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, nil, fmt.Errorf("unexpected gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	if typ == filestore.FileTypeDirectory {
		return f, nil, nil
	}

	rev := &filestore.Revision{
		ID:        uuid.New(),
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
		return nil, nil, fmt.Errorf("unexpected gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return f, rev, nil
}

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

	res := q.db.WithContext(ctx).Raw(`
					SELECT id, path, depth, typ, root_id, created_at, updated_at, mime_type
					FROM filesystem_files
					WHERE root_id=? AND depth=? AND path LIKE ?
					ORDER BY path ASC`,
		q.rootID, filestore.GetPathDepth(path)+1, addTrailingSlash(path)+"%").
		Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

func (q *RootQuery) checkRootExists(ctx context.Context) error {
	zeroUUID := (uuid.UUID{}).String()

	if zeroUUID == q.rootID.String() {
		n := &filestore.Root{}
		res := q.db.WithContext(ctx).Table("filesystem_roots").Where("namespace", q.namespace).First(n)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("root not found, ns: '%s', err: %w", q.namespace, filestore.ErrNotFound)
		}
		if res.Error != nil {
			return res.Error
		}

		q.root = n
		q.rootID = n.ID

		return nil
	}

	n := &filestore.Root{}
	res := q.db.WithContext(ctx).Table("filesystem_roots").Where("id", q.rootID).First(n)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("root not found, id: '%s', err: %w", q.rootID, filestore.ErrNotFound)
	}
	if res.Error != nil {
		return res.Error
	}

	q.root = n

	return nil
}

func (q *RootQuery) SetNamespace(ctx context.Context, namespace string) error {
	res := q.db.WithContext(ctx).Exec(`UPDATE filesystem_roots
		SET namespace = ?
		WHERE id = ?`,
		namespace,
		q.rootID,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return filestore.ErrNotFound
	}

	return nil
}
