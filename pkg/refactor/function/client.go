// nolint
package function

type client interface {
	createService(cfg *Config) error
	updateService(cfg *Config) error
	deleteService(id string) error
	listServices() ([]Status, error)
}

type ClientConfig struct {
	ServiceAccount string `yaml:"service_account"`
	Namespace      string `yaml:"namespace"`
	IngressClass   string `yaml:"ingress_class"`

	Sidecar string `yaml:"sidecar"`

	MaxScale int    `yaml:"max_scale"`
	NetShape string `yaml:"net_shape"`
}
