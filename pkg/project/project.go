package project

const ConfigFile = ".direktiv.yaml"

type Config struct {
	Ignore []string `yaml:"ignore"`
}
