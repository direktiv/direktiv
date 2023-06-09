package datastoresql

import (
	"context"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlServicesStore struct {
	db *gorm.DB
}

func (s *sqlServicesStore) GetByNamespaceAndName(ctx context.Context, namespace uuid.UUID, name string) (*core.Service, error) {
	//TODO implement me
	panic("implement me")
}

func (s *sqlServicesStore) GetOneByUrl(ctx context.Context, url string) (*core.Service, error) {
	//TODO implement me
	panic("implement me")
}

func (s *sqlServicesStore) Create(ctx context.Context, secret *core.Service) error {
	//TODO implement me
	panic("implement me")
}

func (s *sqlServicesStore) Update(ctx context.Context, secret *core.Service) error {
	//TODO implement me
	panic("implement me")
}

func (s *sqlServicesStore) GetAll(ctx context.Context) ([]*core.Service, error) {
	//TODO implement me
	panic("implement me")
}

func (s *sqlServicesStore) DeleteByName(ctx context.Context, name string) error {
	//TODO implement me
	panic("implement me")
}

var _ core.ServicesStore = &sqlServicesStore{}
