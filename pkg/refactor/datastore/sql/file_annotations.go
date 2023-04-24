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
	annotations := &core.FileAnnotations{}
	res := s.db.WithContext(ctx).Table("file_annotations").Where("file_id", fileID).First(annotations)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, core.ErrFileAnnotationsNotSet
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return annotations, nil
}

func (s *sqlFileAnnotationsStore) Set(ctx context.Context, annotations *core.FileAnnotations) error {
	res := s.db.WithContext(ctx).
		Table("file_annotations").
		Where("file_id", annotations.FileID).
		Update("data", annotations.Data)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 1 {
		return fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 1 {
		return nil
	}
	res = s.db.WithContext(ctx).Table("file_annotations").Create(annotations)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

var _ core.FileAnnotationsStore = &sqlFileAnnotationsStore{}
