package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Secret are namespace level variables that are hold sensitive data, can be used inside workflows the same namespace.
type Secret struct {
	Namespace uuid.UUID

	Name string
	Data []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SecretsStore responsible for fetching and setting namespace secrets from datastore.
type SecretsStore interface {
	// Get gets a single namespace secret from the store. if no record found,
	// it returns ErrSecretNotFound error.
	Get(ctx context.Context, namespace uuid.UUID, name string) (*Secret, error)

	// Set either creates (if not exists) a secret or updates the existing one. Param name should be unique.
	Set(ctx context.Context, namespace uuid.UUID, secret *Secret) error

	// GetAll lists all namespace secrets.
	GetAll(ctx context.Context, namespace uuid.UUID) ([]*Secret, error)
}
