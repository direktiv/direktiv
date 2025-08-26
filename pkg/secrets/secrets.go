package secrets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/cache"
	"github.com/direktiv/direktiv/pkg/database"
)

var (
	ErrNotFound = errors.New("ErrNotFound")
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

type Handler struct {
	db    *database.DB
	cache *cache.Cache
}

type Wrapper struct {
	namespace string
	secrets   Secrets
	cache     *cache.Cache
}

func NewHandler(db *database.DB, cache *cache.Cache) *Handler {
	return &Handler{
		db:    db,
		cache: cache,
	}
}

func (sm *Handler) SecretsForNamespace(ctx context.Context, namespace string) (Secrets, error) {
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

func (sw *Wrapper) Get(ctx context.Context, name string) (*Secret, error) {
	value, exists := sw.cache.Get(sw.keyNameforSecret(name))
	if exists {
		s, ok := value.(*Secret)
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

func (sw *Wrapper) Set(ctx context.Context, secret *Secret) (*Secret, error) {
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

func (sw *Wrapper) GetAll(ctx context.Context) ([]*Secret, error) {
	return sw.secrets.GetAll(ctx)
}

func (sw *Wrapper) Update(ctx context.Context, secret *Secret) (*Secret, error) {
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
