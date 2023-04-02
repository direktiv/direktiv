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
	rev          *filestore.Revision
	checksumFunc filestore.CalculateChecksumFunc
	db           *gorm.DB
}

var _ filestore.RevisionQuery = &RevisionQuery{}

func (q *RevisionQuery) Delete(ctx context.Context, force bool) error {
	res := q.db.WithContext(ctx).Delete(&filestore.Revision{}, q.rev.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (q *RevisionQuery) GetData(ctx context.Context) (io.ReadCloser, error) {
	rev := &filestore.Revision{ID: q.rev.ID}
	res := q.db.WithContext(ctx).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}
	reader := bytes.NewReader(rev.Data)
	readCloser := io.NopCloser(reader)

	return readCloser, nil
}

func (q *RevisionQuery) SetData(ctx context.Context, dataReader io.Reader) (*filestore.Revision, error) {
	rev := &filestore.Revision{ID: q.rev.ID}

	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}
	newChecksum := string(q.checksumFunc(data))

	// if same checksum, do nothing, return the same revision.
	if q.rev.Checksum == newChecksum {
		return q.rev, err
	}

	res := q.db.WithContext(ctx).
		Update("data", data).
		Update("checksum", newChecksum).
		First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *RevisionQuery) SetTags(ctx context.Context, tags filestore.RevisionTags) (*filestore.Revision, error) {
	rev := &filestore.Revision{ID: q.rev.ID}
	res := q.db.WithContext(ctx).Update("tags", tags).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}

func (q *RevisionQuery) SetCurrent(ctx context.Context) (*filestore.Revision, error) {
	// set current revisions 'is_current' flag to false.
	res := q.db.WithContext(ctx).
		Model(&filestore.Revision{}).
		Where("file_id", q.rev.FileID).
		Where("is_current", true).
		Update("is_current", false)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// set revision 'is_current' flag to true by id.
	rev := &filestore.Revision{ID: q.rev.ID}
	res = q.db.WithContext(ctx).Update("is_current", true).First(rev)
	if res.Error != nil {
		return nil, res.Error
	}

	return rev, nil
}
