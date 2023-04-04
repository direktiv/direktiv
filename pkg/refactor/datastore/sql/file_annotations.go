package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlFileAnnotationsStore struct {
	db *gorm.DB
}

func (s *sqlFileAnnotationsStore) Get(ctx context.Context, fileID uuid.UUID) (*core.FileAnnotations, error) {
	annotations := &core.FileAnnotations{FileID: fileID}
	res := s.db.WithContext(ctx).First(annotations)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, core.ErrFileAnnotationsNotSet
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return annotations, nil
}

func (s *sqlFileAnnotationsStore) Set(ctx context.Context, annotations *core.FileAnnotations) error {
	res := s.db.WithContext(ctx).Create(annotations)
	if res.Error != nil {
		return s.update(ctx, annotations)
	}

	return nil
}

func (s *sqlFileAnnotationsStore) update(ctx context.Context, annotations *core.FileAnnotations) error {
	res := s.db.WithContext(ctx).
		Model(&core.FileAnnotations{}).
		Where("file_id", annotations.FileID).
		Update("data", annotations.Data)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

var _ core.FileAnnotationsStore = &sqlFileAnnotationsStore{}
