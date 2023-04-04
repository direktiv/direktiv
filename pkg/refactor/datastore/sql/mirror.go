package sql

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlMirrorStore struct {
	db *gorm.DB
}

func (s sqlMirrorStore) UpdateProcess(ctx context.Context, process *mirror.Process) (*mirror.Process, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) CreateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) UpdateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) GetConfig(ctx context.Context, id uuid.UUID) (*mirror.Config, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) GetConfigByNamespace(ctx context.Context, namespace string) (*mirror.Config, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) CreateProcess(ctx context.Context, mirror *mirror.Process) (*mirror.Process, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) GetProcess(ctx context.Context, id uuid.UUID) (*mirror.Process, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) GetProcessesByConfig(ctx context.Context, configID uuid.UUID) ([]*mirror.Process, error) {
	// TODO implement me
	panic("implement me")
}

var _ mirror.Store = sqlMirrorStore{}
