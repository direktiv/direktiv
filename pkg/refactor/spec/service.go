package spec

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type Service struct {
	DirektivAPI string `yaml:"direktiv_api"`
	Image       string `yaml:"image"`
	Cmd         string `yaml:"cmd"`
	Size        string `yaml:"size"`
	Scale       int    `yaml:"scale"`
}

func ParseService(data []byte) (*Service, error) {
	res := &Service{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "service/v1") {
		return nil, errors.New("invalid service api version")
	}

	return res, nil
}

type WorkflowServiceDefinition struct {
	//nolint
	Typ   string `yaml:"type"`
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
	Scale int    `yaml:"scale"`
	Size  string `yaml:"size"`
	Cmd   string `yaml:"cmd"`
}

func ParseWorkflowServiceDefinition(data []byte) (*WorkflowServiceDefinition, error) {
	res := &WorkflowServiceDefinition{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
