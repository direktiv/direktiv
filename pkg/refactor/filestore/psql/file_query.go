package psql

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileQuery struct {
	file *filestore.File
	db   *gorm.DB
}

func (q *FileQuery) GetRevisionByTag(ctx context.Context, tag string) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).
		Where("file_id", q.file.ID).
		Where("tags LIKE ?", "%"+tag+"%").
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *FileQuery) GetRevision(ctx context.Context, id uuid.UUID) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	rev := &filestore.Revision{ID: id}
	res := q.db.WithContext(ctx).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *FileQuery) GetAllRevisions(ctx context.Context) ([]*filestore.Revision, error) {
	//TODO implement me
	//panic("implement me")
	return nil, nil
}

var _ filestore.FileQuery = &FileQuery{}

func (q *FileQuery) Delete(ctx context.Context, force bool) error {
	res := q.db.WithContext(ctx).Delete(&filestore.File{}, q.file.ID)
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
		Where("file_id", q.file.ID).
		Where("is_current", true).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *FileQuery) CreateRevision(ctx context.Context, tags filestore.RevisionTags) (*filestore.Revision, error) {
	if q.file.Typ == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileTypeIsDirectory
	}

	// read file data to copy it to a new revision that we will create.
	dataReader, err := q.GetData(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}

	// set current revisions 'is_current' flag to false.
	res := q.db.WithContext(ctx).
		Model(&filestore.Revision{}).
		Where("file_id", q.file.ID).
		Where("is_current", true).
		Update("is_current", false)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// create a new file revision by copying data.
	newRev := &filestore.Revision{
		ID:   uuid.New(),
		Tags: tags,

		FileID:    q.file.ID,
		IsCurrent: true,

		Data: data,
	}
	res = q.db.WithContext(ctx).Create(newRev)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return newRev, nil
}
