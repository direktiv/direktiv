package datastoresql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlSecretsStore struct {
	db *gorm.DB
}

// nolint
func (s sqlSecretsStore) CreateFolder(ctx context.Context, namespace uuid.UUID, name string) error {
	// TODO: potential un-used feature that we can remove, check with Jens.
	panic("implement me")
}

func (s sqlSecretsStore) Update(ctx context.Context, secret *core.Secret) error {
	res := s.db.WithContext(ctx).Exec(`UPDATE secrets SET data = ? WHERE namespace_id = ? and name = ?`,
		secret.Data, secret.NamespaceID, secret.Name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

// nolint
func (s sqlSecretsStore) DeleteFolder(ctx context.Context, id uuid.UUID, key string) error {
	// TODO: potential un-used feature that we can remove, check with Jens.
	panic("implement me")
}

func (s sqlSecretsStore) Delete(ctx context.Context, namespaceID uuid.UUID, name string) error {
	res := s.db.WithContext(ctx).Exec(`DELETE FROM secrets WHERE namespace_id = ? and name = ?`,
		namespaceID, name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

// nolint
func (s sqlSecretsStore) Search(ctx context.Context, namespace uuid.UUID, name string) ([]*core.Secret, error) {
	// TODO: potential un-used feature that we can remove, check with Jens.
	panic("implement me")
}

func (s sqlSecretsStore) Get(ctx context.Context, namespace uuid.UUID, name string) (*core.Secret, error) {
	secret := &core.Secret{}
	res := s.db.WithContext(ctx).Raw(`
							SELECT id, namespace_id, name, data FROM secrets WHERE "namespace_id" = ? AND name = ?`,
		namespace, name).
		First(secret)
	if res.Error != nil {
		return nil, res.Error
	}

	return secret, nil
}

func (s sqlSecretsStore) Set(ctx context.Context, secret *core.Secret) error {
	res := s.db.WithContext(ctx).Exec(`
							INSERT INTO secrets(id, namespace_id, name, data) VALUES(?, ?, ?, ?);
							`, secret.ID, secret.NamespaceID, secret.Name, secret.Data)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm insert count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s sqlSecretsStore) GetAll(ctx context.Context, namespaceID uuid.UUID) ([]*core.Secret, error) {
	var secrets []*core.Secret

	res := s.db.WithContext(ctx).Raw(`
							SELECT id, namespace_id, name, data FROM secrets WHERE "namespace_id" = ?`,
		namespaceID).
		Find(&secrets)
	if res.Error != nil {
		return nil, res.Error
	}

	return secrets, nil
}

var _ core.SecretsStore = &sqlSecretsStore{}
