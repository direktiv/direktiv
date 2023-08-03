package project

const ConfigFileName = ".direktiv.yaml"

type Config struct {
	Ignore []string `yaml:"ignore"`
}
