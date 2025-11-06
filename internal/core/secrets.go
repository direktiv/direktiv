package core

import (
	"context"
	"time"
)

const SecretsCacheName = "secrets"

type Secret struct {
	Name      string    `json:"name"`
	Data      []byte    `json:"data"`
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

type SecretsManager interface {
	SecretsForNamespace(ctx context.Context, namespace string) (Secrets, error)
}
