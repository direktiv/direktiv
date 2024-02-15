package core

import (
	"context"
	"errors"
	"time"
)

var ErrSecretNotFound = errors.New("secret not found")

// Secret are namespace level variables that are hold sensitive data, can be used inside workflows the same namespace.
type Secret struct {
	Name string `json:"name"`

	Namespace string `json:"-"`

	Data []byte `json:"-"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SecretsStore responsible for fetching and setting namespace secrets from datastore.
type SecretsStore interface {
	// Get gets a single namespace secret from the store. if no record found,
	// it returns ErrSecretNotFound error.
	Get(ctx context.Context, namespace string, name string) (*Secret, error)

	// Set either creates (if not exists) a secret or updates the existing one. Param name should be unique.
	Set(ctx context.Context, secret *Secret) error

	// GetAll lists all namespace secrets.
	GetAll(ctx context.Context, namespace string) ([]*Secret, error)

	// Update changes a secret data.
	Update(ctx context.Context, secret *Secret) error

	// Delete removes a specific secret by name.
	Delete(ctx context.Context, namespace string, name string) error
}
