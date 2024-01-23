package spec

import (
	"errors"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"gopkg.in/yaml.v3"
)

type ServiceFile struct {
	DirektivAPI string                     `yaml:"direktiv_api"`
	Image       string                     `yaml:"image"`
	Cmd         string                     `yaml:"cmd"`
	Size        string                     `yaml:"size"`
	Scale       int                        `yaml:"scale"`
	Envs        []core.EnvironmentVariable `yaml:"envs"`
	Patches     []core.ServicePatch        `yaml:"patches"`
}

func ParseServiceFile(data []byte) (*ServiceFile, error) {
	res := &ServiceFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "service/v1") {
		return nil, errors.New("invalid service api version")
	}

	return res, nil
}
