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
	file filestore.File
	db   *gorm.DB
	//nolint:containedctx
	ctx context.Context
}

var _ filestore.FileQuery = &FileQuery{}

func (q *FileQuery) Delete(force bool) error {
	res := q.db.WithContext(q.ctx).Delete(&File{}, q.file.GetID())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *FileQuery) GetData() (io.ReadCloser, error) {
	if q.file.GetType() == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileIsNotDirectory
	}
	rev := &Revision{FileID: q.file.GetID()}
	res := q.db.WithContext(q.ctx).
		Where("file_id", q.file.GetID()).
		Where("is_current", true).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	reader := bytes.NewReader(rev.Data)
	readCloser := io.NopCloser(reader)

	return readCloser, nil
}

func (q *FileQuery) GetCurrentRevision() (filestore.Revision, error) {
	if q.file.GetType() == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileIsNotDirectory
	}

	rev := &Revision{FileID: q.file.GetID()}
	res := q.db.WithContext(q.ctx).
		Where("file_id", q.file.GetID()).
		Where("is_current", true).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *FileQuery) CreateRevision(tags filestore.RevisionTags) (filestore.Revision, error) {
	if q.file.GetType() == filestore.FileTypeDirectory {
		return nil, filestore.ErrFileIsNotDirectory
	}

	// read file data to copy it to a new revision that we will create.
	dataReader, err := q.GetData()
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}

	// set current revisions 'is_current' flag to false.
	res := q.db.WithContext(q.ctx).
		Model(&Revision{}).
		Where("file_id", q.file.GetID()).
		Where("is_current", true).
		Update("is_current", false)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// create a new file revision by copying data.
	newRev := &Revision{
		ID:   uuid.New(),
		Tags: tags.String(),

		FileID:    q.file.GetID(),
		IsCurrent: true,

		Data: data,
	}
	res = q.db.WithContext(q.ctx).Create(newRev)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return newRev, nil
}
