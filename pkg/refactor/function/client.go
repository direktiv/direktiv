// nolint
package function

type client interface {
	createService(cfg *Config) error
	updateService(cfg *Config) error
	deleteService(id string) error
	listServices() ([]Status, error)
}

type ClientConfig struct {
	ServiceAccount string `yaml:"serviceAccount"`
	Namespace      string `yaml:"namespace"`
	IngressClass   string `yaml:"ingressClass"`

	Sidecar string `yaml:"sidecar"`

	MaxScale int    `yaml:"maxScale"`
	NetShape string `yaml:"netShape"`
}
