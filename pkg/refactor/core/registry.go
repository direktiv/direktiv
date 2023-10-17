package core

type Registry struct {
	Namespace string `json:"namespace"`
	ID        string `json:"id"`
	Url       string `json:"url"`
	User      string `json:"user"`
	Password  string `json:"password,omitempty"`
}

type RegistryManager interface {
	ListRegistries(namespace string) ([]*Registry, error)
	DeleteRegistry(namespace string, id string) error
	StoreRegistry(registry *Registry) (*Registry, error)
}
