package sql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlFileAttributesStore struct {
	db *gorm.DB
}

func (s *sqlFileAttributesStore) Get(ctx context.Context, fileID uuid.UUID) (*core.FileAttributes, error) {
	attrs := &core.FileAttributes{FileID: fileID}
	res := s.db.WithContext(ctx).First(attrs)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, core.ErrFileAttributesNotSet
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return attrs, nil
}

func (s *sqlFileAttributesStore) Set(ctx context.Context, fileAttributes *core.FileAttributes) error {
	res := s.db.WithContext(ctx).Create(fileAttributes)
	if res.Error != nil {
		return s.update(ctx, fileAttributes)
	}

	return nil
}

func (s *sqlFileAttributesStore) update(ctx context.Context, fileAttributes *core.FileAttributes) error {
	res := s.db.WithContext(ctx).
		Model(&core.FileAttributes{}).
		Where("file_id", fileAttributes.FileID).
		Update("value", fileAttributes.Value)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

var _ core.FileAttributesStore = &sqlFileAttributesStore{}
