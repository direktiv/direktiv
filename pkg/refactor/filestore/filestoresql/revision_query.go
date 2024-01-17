package filestoresql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"gorm.io/gorm"
)

type RevisionQuery struct {
	rev *filestore.Revision
	db  *gorm.DB
}

var _ filestore.RevisionQuery = &RevisionQuery{}

func (q *RevisionQuery) Delete(ctx context.Context) error {
	res := q.db.WithContext(ctx).Exec(`DELETE FROM filesystem_revisions WHERE id = ?`, q.rev.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *RevisionQuery) GetData(ctx context.Context) ([]byte, error) {
	rev := &filestore.Revision{}
	res := q.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_revisions 
					WHERE id=?
					`, q.rev.ID).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev.Data, nil
}

func (q *RevisionQuery) SetCurrent(ctx context.Context) (*filestore.Revision, error) {
	// set current revisions 'is_current' flag to false.
	res := q.db.WithContext(ctx).Exec(`
				UPDATE filesystem_revisions 
				SET is_current=false
				WHERE is_current=true AND file_id=?
				`, q.rev.FileID)
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
