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
	rev filestore.Revision
	db  *gorm.DB
	ctx context.Context
}

var _ filestore.RevisionQuery = &RevisionQuery{}

func (q *RevisionQuery) Delete(force bool) error {
	res := q.db.WithContext(q.ctx).Delete(&Revision{}, q.rev.GetID())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *RevisionQuery) GetData() (io.ReadCloser, error) {
	rev := &Revision{ID: q.rev.GetID()}
	res := q.db.WithContext(q.ctx).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	reader := bytes.NewReader(rev.Data)
	readCloser := io.NopCloser(reader)

	return readCloser, nil
}

func (q *RevisionQuery) SetTags(tags filestore.RevisionTags) (filestore.Revision, error) {
	rev := &Revision{ID: q.rev.GetID()}
	res := q.db.WithContext(q.ctx).Update("tags", tags.String()).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *RevisionQuery) SetCurrent() (filestore.Revision, error) {
	// set current revisions 'is_current' flag to false.
	res := q.db.WithContext(q.ctx).
		Model(&Revision{}).
		Where("file_id", q.rev.GetFileID()).
		Where("is_current", true).
		Update("is_current", false)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// set revision 'is_current' flag to true by id.
	rev := &Revision{ID: q.rev.GetID()}
	res = q.db.WithContext(q.ctx).Update("is_current", true).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}
