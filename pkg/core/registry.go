package core

import "time"

type Registry struct {
	Namespace string    `json:"namespace"`
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	User      string    `json:"user"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type RegistryManager interface {
	ListRegistries(namespace string) ([]*Registry, error)
	DeleteRegistry(namespace string, id string) error
	DeleteNamespace(namespace string) error
	StoreRegistry(registry *Registry) (*Registry, error)
	TestLogin(registry *Registry) error
}
