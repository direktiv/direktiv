package datastoresql

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlSecretsStore struct {
	db *gorm.DB
}

func (s sqlSecretsStore) Get(ctx context.Context, namespace uuid.UUID, name string) (*core.Secret, error) {
	//TODO implement me
	panic("implement me")
}

func (s sqlSecretsStore) Set(ctx context.Context, namespace uuid.UUID, secret *core.Secret) error {
	//TODO implement me
	panic("implement me")
}

func (s sqlSecretsStore) GetAll(ctx context.Context, namespace uuid.UUID) ([]*core.Secret, error) {
	//TODO implement me
	panic("implement me")
}

var _ core.SecretsStore = &sqlSecretsStore{}
