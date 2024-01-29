package datastoresql

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlNamespacesStore struct {
	db *gorm.DB
}

func (s *sqlNamespacesStore) GetByID(ctx context.Context, id uuid.UUID) (*core.Namespace, error) {
	namespace := &core.Namespace{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, name, config, created_at, updated_at 
							FROM namespaces 
							WHERE id=?`,
		id).
		First(namespace)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return namespace, nil
}

func (s *sqlNamespacesStore) GetByName(ctx context.Context, name string) (*core.Namespace, error) {
	namespace := &core.Namespace{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, name, config, created_at, updated_at 
							FROM namespaces 
							WHERE name=?`,
		name).
		First(namespace)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return namespace, nil
}

func (s *sqlNamespacesStore) GetAll(ctx context.Context) ([]*core.Namespace, error) {
	var namespaces []*core.Namespace
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, name, config, created_at, updated_at
							FROM namespaces`).
		Find(&namespaces)
	if res.Error != nil {
		return nil, res.Error
	}

	return namespaces, nil
}

func (s *sqlNamespacesStore) Update(ctx context.Context, namespace *core.Namespace) (*core.Namespace, error) {
	res := s.db.WithContext(ctx).Exec(`
						UPDATE namespaces
						SET
							 name=?, config=?
						WHERE id=?`,
		namespace.Name, namespace.Config, namespace.ID)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected > 1 {
		return nil, fmt.Errorf("unexpected namespaces update count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 0 {
		return nil, datastore.ErrNotFound
	}

	return s.GetByID(ctx, namespace.ID)
}

func (s *sqlNamespacesStore) Delete(ctx context.Context, name string) error {
	res := s.db.WithContext(ctx).Exec(`DELETE FROM namespaces WHERE  name=?`, name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 1 {
		return fmt.Errorf("unexpected namespaces delete count, got: %d, want: %d", res.RowsAffected, 1)
	}
	if res.RowsAffected == 0 {
		return datastore.ErrNotFound
	}

	return nil
}

func (s *sqlNamespacesStore) Create(ctx context.Context, namespace *core.Namespace) (*core.Namespace, error) {
	const nameRegex = `^(([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))$`
	matched, _ := regexp.MatchString(nameRegex, namespace.Name)
	if !matched {
		return nil, core.ErrInvalidNamespaceName
	}

	newUUID := uuid.New()
	res := s.db.WithContext(ctx).Exec(`
							INSERT INTO namespaces(id, name, config) VALUES(?, ?, ?);
							`, newUUID, namespace.Name, namespace.Config)

	if res.Error != nil && strings.Contains(res.Error.Error(), "duplicate key") {
		return nil, core.ErrDuplicatedNamespaceName
	}
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected namespaces insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetByID(ctx, newUUID)
}

var _ core.NamespacesStore = &sqlNamespacesStore{}
