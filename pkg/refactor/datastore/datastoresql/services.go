package datastoresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlServicesStore struct {
	db *gorm.DB
}

func (s *sqlServicesStore) GetByNamespaceIDAndName(ctx context.Context, namespaceID uuid.UUID, name string) (*core.Service, error) {
	service := &core.Service{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, namespace_id, name, url, data, created_at, updated_at 
							FROM services 
							WHERE namespace_id=? AND name=?`,
		namespaceID, name).
		First(service)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return service, nil
}

func (s *sqlServicesStore) GetOneByURL(ctx context.Context, url string) (*core.Service, error) {
	service := &core.Service{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, namespace_id, name, url, data, created_at, updated_at 
							FROM services 
							WHERE url=?`,
		url).
		First(service)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return service, nil
}

func (s *sqlServicesStore) Create(ctx context.Context, service *core.Service) error {
	newUUID := uuid.New()
	res := s.db.WithContext(ctx).Exec(`
							INSERT INTO services(id, namespace_id, name, url, data) VALUES(?, ?, ?, ?, ?);
							`, newUUID, service.NamespaceID, service.Name, service.URL, service.Data)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected services insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s *sqlServicesStore) Update(ctx context.Context, service *core.Service) error {
	res := s.db.WithContext(ctx).Exec(`
						UPDATE services
						SET
							 url=?, data=?
						WHERE id=?`,
		service.URL, service.Data, service.ID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 1 {
		return fmt.Errorf("unexpedted services update count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 0 {
		return datastore.ErrNotFound
	}

	return nil
}

func (s *sqlServicesStore) GetAll(ctx context.Context) ([]*core.Service, error) {
	var services []*core.Service
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, namespace_id, name, url, data, created_at, updated_at 
							FROM services`).
		Find(&services)
	if res.Error != nil {
		return nil, res.Error
	}

	return services, nil
}

func (s *sqlServicesStore) DeleteByName(ctx context.Context, name string) error {
	res := s.db.WithContext(ctx).Exec(`DELETE FROM services WHERE  name=?`, name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 1 {
		return fmt.Errorf("unexpedted services delete count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 0 {
		return datastore.ErrNotFound
	}

	return nil
}

var _ core.ServicesStore = &sqlServicesStore{}
