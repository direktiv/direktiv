package api

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/model"
	"gopkg.in/yaml.v3"
)

const (
	FiltersAPIV1 = "filters/v1"
)

type Filter struct {
	Name             string `yaml:"name"`
	InlineJavascript string `yaml:"inline_javascript"`
	Source           string `yaml:"source"`
}

type Filters struct {
	DirektivAPI string   `yaml:"direktiv_api"`
	Filters     []Filter `yaml:"filters"`
}

const (
	ServicesAPIV1 = "services/v1"
)

type Service struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

type Services struct {
	DirektivAPI string    `yaml:"direktiv_api"`
	Services    []Service `yaml:"services"`
}

const (
	WorkflowAPIV1 = "workflow/v1"
)

func LoadResource(data []byte) (interface{}, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("error parsing direktiv resource: %w", err)
	}

	x, exists := m["direktiv_api"]
	if !exists {
		return nil, errors.New("error parsing direktiv resource: missing 'direktiv_api'")
	}

	s, ok := x.(string)
	if !ok {
		return nil, errors.New("error parsing direktiv resource: invalid 'direktiv_api'")
	}

	switch s {
	case FiltersAPIV1:
		filters := new(Filters)
		err = yaml.Unmarshal(data, &filters)
		if err != nil {
			return nil, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return filters, nil
	case ServicesAPIV1:
		services := new(Services)
		err = yaml.Unmarshal(data, &services)
		if err != nil {
			return nil, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return services, nil
	case WorkflowAPIV1:
		wf := new(model.Workflow)
		err = wf.Load(data)
		if err != nil {
			return nil, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return wf, nil
	default:
		return nil, fmt.Errorf("error parsing direktiv resource: invalid 'direktiv_api': \"%s\"", s)
	}
}
