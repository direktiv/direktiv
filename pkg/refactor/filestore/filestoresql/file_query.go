package filestoresql

import (
	"context"
	"fmt"
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
	rq := &RootQuery{
		rootID:       q.file.RootID,
		db:           q.db,
		checksumFunc: filestore.DefaultCalculateChecksum,
	}

	children, err := rq.ReadDirectory(ctx, q.file.Path)
	if err != nil {
		return err
	}

	for _, child := range children {
		cq := &FileQuery{
			file:         child,
			db:           q.db,
			checksumFunc: filestore.DefaultCalculateChecksum,
		}

		err = cq.setPath(ctx, filepath.Join(path, child.Name()))
		if err != nil {
			return err
		}
	}

	err = q.setPathForFileType(ctx, path)
	if err != nil {
		return err
	}

	return nil
}

func (q *FileQuery) setPath(ctx context.Context, path string) error {
	if q.file.Typ == filestore.FileTypeDirectory {
		return q.setPathForDirectoryType(ctx, path)
	}

	return q.setPathForFileType(ctx, path)
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

	err = q.setPath(ctx, path)
	if err != nil {
		return err
	}

	return nil
}

func (q *FileQuery) GetRevision(ctx context.Context) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).Raw(`
							SELECT * FROM filesystem_revisions WHERE "file_id" = ?`,
		q.file.ID).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

var _ filestore.FileQuery = &FileQuery{}

//nolint:revive
func (q *FileQuery) Delete(ctx context.Context, force bool) error {
	res := q.db.WithContext(ctx).Exec(`DELETE FROM filesystem_files WHERE id = ?`, q.file.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *FileQuery) GetData(ctx context.Context) ([]byte, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}
	rev := &filestore.Revision{}

	res := q.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_revisions
					WHERE file_id=?
					`, q.file.ID).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev.Data, nil
}

func (q *FileQuery) CreateRevision(ctx context.Context, data []byte) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	newChecksum := string(q.checksumFunc(data))

	// create a new file revision.
	newRev := &filestore.Revision{
		ID:     uuid.New(),
		FileID: q.file.ID,

		Checksum: newChecksum,
		Data:     data,
	}
	res := q.db.WithContext(ctx).Table("filesystem_revisions").Create(newRev)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return newRev, nil
}
