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
	// UpdatedAt time.Time `json:"updatedAt"`
}

type SecretsManager interface {
	Create(ctx context.Context, namespace string, secret *Secret) (*Secret, error)
	Get(ctx context.Context, namespace, name string) (*Secret, error)
	GetAll(ctx context.Context, namespace string) ([]*Secret, error)
	Update(ctx context.Context, namespace string, secret *Secret) (*Secret, error)
	Delete(ctx context.Context, namespace, name string) error
	DeleteForNamespace(ctx context.Context, namespace string) error
}
