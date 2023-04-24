package psql

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"gorm.io/gorm"
)

type RevisionQuery struct {
	rev *filestore.Revision
	db  *gorm.DB
}

var _ filestore.RevisionQuery = &RevisionQuery{}

//nolint:revive
func (q *RevisionQuery) Delete(ctx context.Context, force bool) error {
	res := q.db.WithContext(ctx).Table("filesystem_revisions").Delete(&filestore.Revision{}, q.rev.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *RevisionQuery) GetData(ctx context.Context) (io.ReadCloser, error) {
	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).Table("filesystem_revisions").Where("id", q.rev.ID).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	reader := bytes.NewReader(rev.Data)
	readCloser := io.NopCloser(reader)

	return readCloser, nil
}

func (q *RevisionQuery) SetTags(ctx context.Context, tags filestore.RevisionTags) (*filestore.Revision, error) {
	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).Table("filesystem_revisions").
		Where("id", q.rev.ID).
		Update("tags", tags).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return rev, nil
}

func (q *RevisionQuery) SetCurrent(ctx context.Context) (*filestore.Revision, error) {
	// set current revisions 'is_current' flag to false.
	res := q.db.WithContext(ctx).Table("filesystem_revisions").
		Where("file_id", q.rev.FileID).
		Where("is_current", true).
		Update("is_current", false)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// set revision 'is_current' flag to true by id.
	rev := &filestore.Revision{}
	res = q.db.WithContext(ctx).Table("filesystem_revisions").Update("is_current", true).Where("id", q.rev.ID).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}
