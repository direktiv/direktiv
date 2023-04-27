package psql

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileQuery struct {
	file         *filestore.File
	checksumFunc filestore.CalculateChecksumFunc
	db           *gorm.DB
}

func (q *FileQuery) setPathForFileType(ctx context.Context, path string) error {
	res := q.db.WithContext(ctx).Exec("UPDATE filesystem_files SET path = ?, depth = ? WHERE root_id = ? AND path = ?",
		path, filestore.GetPathDepth(path), q.file.RootID, q.file.Path)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *FileQuery) setPathForDirectoryType(ctx context.Context, path string) error {
	// In SQL, REPLACE(str, param1, param2) function does replace all the occurrences of param1 with param2, this will
	// result in a bug where paths with repetitive components like '/a/b/a/b/a/b'
	// get updated to '/z/b/z/b/z/b' instead of '/z/b/a/b/a/b' when path '/a' get set to '/z'.
	// To overcome this problem with REPLACE(), we prefix REPLACE() parameter with "//" string.

	res := q.db.WithContext(ctx).Exec(`
							UPDATE filesystem_files SET path = REPLACE( "//" || path, "//" || ?, ?), depth = ?
							             WHERE (root_id = ? AND path = ?)
							             OR (root_id = ? AND path LIKE ?)`,
		q.file.Path, path, filestore.GetPathDepth(path),
		q.file.RootID, q.file.Path,
		q.file.RootID, q.file.Path+"/%",
	)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *FileQuery) SetPath(ctx context.Context, path string) error {
	path, err := filestore.SanitizePath(path)
	if err != nil {
		return filestore.ErrInvalidPathParameter
	}

	// check if new path doesn't exist.
	count := 0
	tx := q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ? AND path = ?", q.file.RootID, path).Scan(&count)
	if tx.Error != nil {
		return tx.Error
	}
	if count > 0 {
		return filestore.ErrPathAlreadyExists
	}

	// check if parent dir of the new path exist.
	parentDir := filepath.Dir(path)
	if parentDir != "/" {
		count = 0
		tx = q.db.WithContext(ctx).Raw("SELECT count(id) FROM filesystem_files WHERE root_id = ? AND typ = ? AND path = ?", q.file.RootID, filestore.FileTypeDirectory, parentDir).Scan(&count)
		if tx.Error != nil {
			return tx.Error
		}
		if count == 0 {
			return filestore.ErrNoParentDirectory
		}
	}

	if q.file.Typ == filestore.FileTypeDirectory {
		return q.setPathForDirectoryType(ctx, path)
	}

	return q.setPathForFileType(ctx, path)
}

func (q *FileQuery) getRevisionByID(ctx context.Context, id uuid.UUID) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).Table("filesystem_revisions").
		Where("id", id).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *FileQuery) GetRevision(ctx context.Context, reference string) (*filestore.Revision, error) {
	if reference == "" {
		return nil, filestore.ErrNotFound
	}
	if reference == filestore.Latest {
		return q.GetCurrentRevision(ctx)
	}
	id, err := uuid.Parse(reference)
	if err == nil {
		return q.getRevisionByID(ctx, id)
	}

	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).Raw(`
							SELECT * FROM filesystem_revisions WHERE "file_id" = ? AND
							("tags" = ? OR "tags" LIKE ? OR "tags" LIKE ? OR "tags" LIKE ?)`,
		q.file.ID,
		reference,
		reference+",%",
		"%,"+reference,
		"%,"+reference+",%").
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

//nolint:revive
func (q *FileQuery) GetAllRevisions(ctx context.Context) ([]*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	var list []*filestore.Revision
	res := q.db.WithContext(ctx).
		Table("filesystem_revisions").
		Where("file_id", q.file.ID).
		Order("created_at desc").
		Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	return list, nil
}

var _ filestore.FileQuery = &FileQuery{}

//nolint:revive
func (q *FileQuery) Delete(ctx context.Context, force bool) error {
	res := q.db.WithContext(ctx).Table("filesystem_files").Delete(&filestore.File{}, q.file.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *FileQuery) GetData(ctx context.Context) (io.ReadCloser, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}
	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).
		Table("filesystem_revisions").
		Where("file_id", q.file.ID).
		Where("is_current", true).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	reader := bytes.NewReader(rev.Data)
	readCloser := io.NopCloser(reader)

	return readCloser, nil
}

func (q *FileQuery) GetCurrentRevision(ctx context.Context) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).
		Table("filesystem_revisions").
		Where("file_id", q.file.ID).
		Where("is_current", true).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *FileQuery) CreateRevision(ctx context.Context, tags filestore.RevisionTags, dataReader io.Reader) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}
	newChecksum := string(q.checksumFunc(data))

	// set current revisions 'is_current' flag to false.
	res := q.db.WithContext(ctx).
		Table("filesystem_revisions").
		Where("file_id", q.file.ID).
		Where("is_current", true).
		Update("is_current", false)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// create a new file revision.
	newRev := &filestore.Revision{
		ID:   uuid.New(),
		Tags: tags,

		FileID:    q.file.ID,
		IsCurrent: true,

		Checksum: newChecksum,
		Data:     data,
	}
	res = q.db.WithContext(ctx).Table("filesystem_revisions").Create(newRev)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return newRev, nil
}
