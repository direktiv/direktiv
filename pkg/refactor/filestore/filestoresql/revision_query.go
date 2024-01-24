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
