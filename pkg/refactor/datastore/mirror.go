package datastore

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlMirrorStore struct {
	db *gorm.DB
}

func (s sqlMirrorStore) CreateSettings(ctx context.Context, settings *mirror.Settings) error {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) GetSetting(ctx context.Context, id uuid.UUID) (*mirror.Settings, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) DeleteSetting(ctx context.Context, id uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) CreateActivity(ctx context.Context, activity *mirror.Activity) error {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) GetActivity(ctx context.Context, id uuid.UUID) (*mirror.Activity, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlMirrorStore) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}

var _ mirror.Store = sqlMirrorStore{}
