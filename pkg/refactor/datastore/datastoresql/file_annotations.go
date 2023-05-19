package datastoresql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlFileAnnotationsStore struct {
	db *gorm.DB
}

// Get gets file annotations information of a file.
func (s *sqlFileAnnotationsStore) Get(ctx context.Context, fileID uuid.UUID) (*core.FileAnnotations, error) {
	rawAnnotations := &struct {
		FileID uuid.UUID
		Data   []byte

		CreatedAt time.Time
		UpdatedAt time.Time
	}{}

	res := s.db.WithContext(ctx).Table("file_annotations").Where("file_id", fileID).First(rawAnnotations)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, core.ErrFileAnnotationsNotSet
	}
	if res.Error != nil {
		return nil, res.Error
	}

	data := map[string]string{}
	if err := json.Unmarshal(rawAnnotations.Data, &data); err != nil {
		return nil, err
	}

	return &core.FileAnnotations{
		FileID:    rawAnnotations.FileID,
		Data:      data,
		CreatedAt: rawAnnotations.CreatedAt,
		UpdatedAt: rawAnnotations.UpdatedAt,
	}, nil
}

// Set either creates (if not exists) file annotation information or updates the existing one.
func (s *sqlFileAnnotationsStore) Set(ctx context.Context, annotations *core.FileAnnotations) error {
	data, err := json.Marshal(annotations.Data)
	if err != nil {
		return err
	}

	res := s.db.WithContext(ctx).Table("file_annotations").
		Where("file_id", annotations.FileID).
		Update("data", data)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 1 {
		return fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 1 {
		return nil
	}
	res = s.db.WithContext(ctx).Exec(`
							INSERT INTO file_annotations(file_id, data) VALUES(?, ?);
							`, annotations.FileID, data)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

var _ core.FileAnnotationsStore = &sqlFileAnnotationsStore{}
