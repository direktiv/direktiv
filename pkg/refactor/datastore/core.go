package datastore

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlNamespaceStore struct {
	db *gorm.DB
}

func (s sqlNamespaceStore) Create(ctx context.Context, namespace *core.Namespace) error {
	// TODO implement me
	panic("implement me")
}

func (s sqlNamespaceStore) Get(ctx context.Context, id uuid.UUID) (*core.Namespace, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlNamespaceStore) GetByName(ctx context.Context, name string) (*core.Namespace, error) {
	// TODO implement me
	panic("implement me")
}

func (s sqlNamespaceStore) Delete(ctx context.Context, name string) error {
	// TODO implement me
	panic("implement me")
}

var _ core.NamespaceStore = sqlNamespaceStore{}
