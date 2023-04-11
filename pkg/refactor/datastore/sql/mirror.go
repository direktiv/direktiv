package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlMirrorStore struct {
	db *gorm.DB
}

//nolint:unused
func (s sqlMirrorStore) SetNamespaceVariable(ctx context.Context, namespaceID uuid.UUID, key string, data []byte, hash string, mType string) error {
	// TODO: implement me.
	return nil
}

//nolint:unused
func (s sqlMirrorStore) SetWorkflowVariable(ctx context.Context, workflowID uuid.UUID, key string, data []byte, hash string, mType string) error {
	// TODO: implement me.
	return nil
}

func (s sqlMirrorStore) CreateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	newConfig := *config
	res := s.db.WithContext(ctx).Table("mirror_configs").Create(&newConfig)
	if res.Error != nil {
		return nil, res.Error
	}

	return &newConfig, nil
}

func (s sqlMirrorStore) UpdateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	res := s.db.WithContext(ctx).
		Table("mirror_configs").
		Where("id", config.ID).
		Updates(config)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetConfig(ctx, config.ID)
}

func (s sqlMirrorStore) GetConfig(ctx context.Context, id uuid.UUID) (*mirror.Config, error) {
	config := &mirror.Config{ID: id}
	res := s.db.WithContext(ctx).Table("mirror_configs").First(config)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, mirror.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return config, nil
}

func (s sqlMirrorStore) CreateProcess(ctx context.Context, process *mirror.Process) (*mirror.Process, error) {
	newProcess := *process
	res := s.db.WithContext(ctx).Table("mirror_processes").Create(&newProcess)
	if res.Error != nil {
		return nil, res.Error
	}

	return &newProcess, nil
}

func (s sqlMirrorStore) UpdateProcess(ctx context.Context, process *mirror.Process) (*mirror.Process, error) {
	res := s.db.WithContext(ctx).
		Table("mirror_processes").
		Model(&mirror.Process{}).
		Where("id", process.ID).
		Updates(process)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetProcess(ctx, process.ID)
}

func (s sqlMirrorStore) GetProcess(ctx context.Context, id uuid.UUID) (*mirror.Process, error) {
	process := &mirror.Process{ID: id}
	res := s.db.WithContext(ctx).Table("mirror_processes").First(process)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, mirror.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) GetProcessesByConfig(ctx context.Context, configID uuid.UUID) ([]*mirror.Process, error) {
	var process []*mirror.Process

	res := s.db.WithContext(ctx).
		Table("mirror_processes").
		Where("config_id", configID).
		Find(&process)
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

var _ mirror.Store = sqlMirrorStore{}
