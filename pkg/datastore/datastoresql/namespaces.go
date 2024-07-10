package datastoresql

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlNamespacesStore struct {
	db *gorm.DB
}

func (s *sqlNamespacesStore) GetByID(ctx context.Context, id uuid.UUID) (*datastore.Namespace, error) {
	namespace := &datastore.Namespace{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, name, created_at, updated_at 
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

func (s *sqlNamespacesStore) GetByName(ctx context.Context, name string) (*datastore.Namespace, error) {
	namespace := &datastore.Namespace{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, name, created_at, updated_at 
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

func (s *sqlNamespacesStore) GetAll(ctx context.Context) ([]*datastore.Namespace, error) {
	var namespaces []*datastore.Namespace
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, name, created_at, updated_at
							FROM namespaces ORDER BY created_at DESC `).
		Find(&namespaces)
	if res.Error != nil {
		return nil, res.Error
	}

	return namespaces, nil
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

func (s *sqlNamespacesStore) Create(ctx context.Context, namespace *datastore.Namespace) (*datastore.Namespace, error) {
	const nameRegex = `^(([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))$`
	matched, _ := regexp.MatchString(nameRegex, namespace.Name)
	if !matched {
		return nil, datastore.ErrInvalidNamespaceName
	}

	newUUID := uuid.New()
	res := s.db.WithContext(ctx).Exec(`
							INSERT INTO namespaces(id, name) VALUES(?, ?);
							`, newUUID, namespace.Name)

	if res.Error != nil && strings.Contains(res.Error.Error(), "duplicate key") {
		return nil, datastore.ErrDuplicatedNamespaceName
	}
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected namespaces insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetByID(ctx, newUUID)
}

var _ datastore.NamespacesStore = &sqlNamespacesStore{}
