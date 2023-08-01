package core

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSecretNotFound = errors.New("secret not found")
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
	// TODO: potential un-used feature that we can remove, check with Jens.
	Search(ctx context.Context, namespaceID uuid.UUID, name string) ([]*Secret, error)

	// CreateFolder creates a new secret folder.
	// TODO: potential un-used feature that we can remove, check with Jens.
	CreateFolder(ctx context.Context, namespaceID uuid.UUID, name string) error

	// DeleteFolder removes the whole secrets folder by name.
	// TODO: potential un-used feature that we can remove, check with Jens.
	DeleteFolder(ctx context.Context, namespaceID uuid.UUID, name string) error

	// Update changes a secret data.
	Update(ctx context.Context, secret *Secret) error

	// Delete removes a specific secret by name.
	Delete(ctx context.Context, namespaceID uuid.UUID, name string) error
}
