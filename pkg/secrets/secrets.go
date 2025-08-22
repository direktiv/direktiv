package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/cache"
	"github.com/direktiv/direktiv/pkg/database"
)

type Secret struct {
	Name string `json:"name"`

	Data []byte `json:"data"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Secrets interface {
	Get(ctx context.Context, name string) (*Secret, error)
	Set(ctx context.Context, secret *Secret) (*Secret, error)
	GetAll(ctx context.Context) ([]*Secret, error)
	Update(ctx context.Context, secret *Secret) (*Secret, error)
	Delete(ctx context.Context, name string) error
}

type SecretsHandler struct {
	db    *database.DB
	cache *cache.Cache
}

type SecretsWrapper struct {
	namespace string
	secrets   Secrets
	cache     *cache.Cache
}

func NewSecretsHandler(db *database.DB, cache *cache.Cache) *SecretsHandler {
	return &SecretsHandler{
		db:    db,
		cache: cache,
	}
}

func (sm *SecretsHandler) SecretsForNamespace(namespace string) Secrets {
	dbs := &DBSecrets{
		namespace: namespace,
		db:        sm.db,
	}

	return &SecretsWrapper{
		secrets:   dbs,
		cache:     sm.cache,
		namespace: namespace,
	}
}

func (sw *SecretsWrapper) Get(ctx context.Context, name string) (*Secret, error) {
	var secret *Secret
	value, exists := sw.cache.Get(sw.keyNameforSecret(name))
	if exists {
		err := json.Unmarshal(value, &secret)
		// if no error we return the value, otherwise fetching from implementation
		if err == nil {
			return secret, nil
		}
	}

	s, err := sw.secrets.Get(ctx, name)
	if err == nil {
		sw.addToCache(s)
	}

	return s, err
}

func (sw *SecretsWrapper) Set(ctx context.Context, secret *Secret) (*Secret, error) {
	// set in implementation first
	v, err := sw.secrets.Set(ctx, secret)
	if err != nil {
		return nil, err
	}

	secret.CreatedAt = v.CreatedAt
	secret.UpdatedAt = v.UpdatedAt

	sw.addToCache(secret)

	return v, err
}

func (sw *SecretsWrapper) GetAll(ctx context.Context) ([]*Secret, error) {
	return sw.secrets.GetAll(ctx)
}

func (sw *SecretsWrapper) Update(ctx context.Context, secret *Secret) (*Secret, error) {
	s, err := sw.secrets.Update(ctx, secret)
	sw.addToCache(s)

	return s, err
}

func (sw *SecretsWrapper) Delete(ctx context.Context, name string) error {
	sw.cache.Delete(sw.keyNameforSecret(name))
	return sw.secrets.Delete(ctx, name)
}

func (sw *SecretsWrapper) keyNameforSecret(name string) string {
	return fmt.Sprintf("secret-%s-%s", sw.namespace, name)
}

func (sw *SecretsWrapper) addToCache(secret *Secret) {
	b, err := json.Marshal(secret)
	if err != nil {
		slog.Error("error caching secret", slog.Any("error", err))
	}

	sw.cache.Set(sw.keyNameforSecret(secret.Name), b)
}
