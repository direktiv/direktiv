package secrets

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/internal/cache"
	"github.com/direktiv/direktiv/internal/core"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("ErrNotFound")

type Manager struct {
	db    *gorm.DB
	cache cache.Cache
}

type Wrapper struct {
	namespace string
	secrets   core.Secrets
	cache     cache.Cache
}

func NewManager(db *gorm.DB, cache cache.Cache) core.SecretsManager {
	return &Manager{
		db:    db,
		cache: cache,
	}
}

func (sm *Manager) SecretsForNamespace(ctx context.Context, namespace string) (core.Secrets, error) {
	dbs := &DBSecrets{
		namespace: namespace,
		db:        sm.db,
	}

	return &Wrapper{
		secrets:   dbs,
		cache:     sm.cache,
		namespace: namespace,
	}, nil
}

func (sw *Wrapper) Get(ctx context.Context, name string) (*core.Secret, error) {
	value, exists := sw.cache.Get(sw.keyNameforSecret(name))
	if exists {
		s, ok := value.(*core.Secret)
		if ok {
			return s, nil
		}
	}

	s, err := sw.secrets.Get(ctx, name)
	if err == nil {
		sw.cache.Set(sw.keyNameforSecret(name), s)
	}

	return s, err
}

func (sw *Wrapper) Set(ctx context.Context, secret *core.Secret) (*core.Secret, error) {
	// set in implementation first
	v, err := sw.secrets.Set(ctx, secret)
	if err != nil {
		return nil, err
	}

	secret.CreatedAt = v.CreatedAt
	secret.UpdatedAt = v.UpdatedAt

	sw.cache.Set(sw.keyNameforSecret(secret.Name), secret)

	return v, err
}

func (sw *Wrapper) GetAll(ctx context.Context) ([]*core.Secret, error) {
	return sw.secrets.GetAll(ctx)
}

func (sw *Wrapper) Update(ctx context.Context, secret *core.Secret) (*core.Secret, error) {
	s, err := sw.secrets.Update(ctx, secret)
	if err != nil {
		return nil, err
	}
	sw.cache.Set(sw.keyNameforSecret(secret.Name), s)

	return s, err
}

func (sw *Wrapper) Delete(ctx context.Context, name string) error {
	sw.cache.Delete(sw.keyNameforSecret(name))
	return sw.secrets.Delete(ctx, name)
}

func (sw *Wrapper) keyNameforSecret(name string) string {
	return fmt.Sprintf("secret-%s-%s", sw.namespace, name)
}
