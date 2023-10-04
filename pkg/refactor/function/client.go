// nolint
package function

type client interface {
	createService(cfg *Config) error
	updateService(cfg *Config) error
	deleteService(id string) error
	listServices() ([]Status, error)
}

type ClientConfig struct {
	ServiceAccount string `yaml:"service-account"`
	Namespace      string `yaml:"namespace"`
	IngressClass   string `yaml:"ingress-class"`

	Sidecar string `yaml:"sidecar"`

	MaxScale int    `yaml:"max-scale"`
	NetShape string `yaml:"net-shape"`
}
