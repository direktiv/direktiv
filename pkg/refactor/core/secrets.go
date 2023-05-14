package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Secret are namespace level variables that are hold sensitive data, can be used inside workflows the same namespace.
type Secret struct {
	ID          uuid.UUID
	NamespaceID uuid.UUID

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
	Set(ctx context.Context, secret *Secret) error

	// GetAll lists all namespace secrets.
	GetAll(ctx context.Context, namespaceID uuid.UUID) ([]*Secret, error)

	// Search filter out secretes by name.
	Search(ctx context.Context, namespaceID uuid.UUID, name string) ([]*Secret, error)

	// CreateFolder creates a new secret folder.
	CreateFolder(ctx context.Context, namespaceID uuid.UUID, name string) error

	// Update changes a secret data.
	Update(ctx context.Context, secret *Secret) error

	// DeleteFolder removes the whole secrets folder by name.
	DeleteFolder(ctx context.Context, namespaceID uuid.UUID, key string) error

	// Delete removes a specific secret by name.
	Delete(ctx context.Context, namespaceID uuid.UUID, key string) error
}
