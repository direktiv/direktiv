package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
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
	WorkflowAPIV1 = "workflow/v1"
)

const (
	ServiceAPIV1 = "service/v1"
)

const (
	EndpointAPIV1 = "endpoint/v1"
)

const (
	ConsumerAPIV1 = "consumer/v1"
)

var ErrNotDirektivAPIResource = errors.New("not a direktiv_api resource")

func LoadResource(data []byte) (interface{}, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotDirektivAPIResource, err)
	}

	x, exists := m["direktiv_api"]
	if !exists {
		return nil, fmt.Errorf("%w: missing 'direktiv_api' field", ErrNotDirektivAPIResource)
	}

	s, ok := x.(string)
	if !ok {
		return nil, fmt.Errorf("%w: invalid 'direktiv_api' field", ErrNotDirektivAPIResource)
	}

	switch s {
	case FiltersAPIV1:
		filters := new(Filters)
		err = yaml.Unmarshal(data, &filters)
		if err != nil {
			return &Filters{
				DirektivAPI: s,
			}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return filters, nil

	case WorkflowAPIV1:
		wf := new(Workflow)
		err = wf.Load(data)
		if err != nil {
			return &Workflow{
				DirektivAPI: s,
			}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return wf, nil

	case ServiceAPIV1:
		sf := new(core.ServiceFile)
		err = yaml.Unmarshal(data, &sf)
		if err != nil {
			return &core.ServiceFile{
				DirektivAPI: s,
			}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return sf, nil

	case EndpointAPIV1:
		ef := new(core.EndpointFile)
		err = yaml.Unmarshal(data, &ef)
		if err != nil {
			return &core.EndpointFile{
				DirektivAPI: s,
			}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return ef, nil

	case ConsumerAPIV1:
		ef := new(core.ConsumerFile)
		err = yaml.Unmarshal(data, &ef)
		if err != nil {
			return &core.ConsumerFile{
				DirektivAPI: s,
			}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return ef, nil

	default:
		return nil, fmt.Errorf("error parsing direktiv resource: invalid 'direktiv_api': \"%s\"", s)
	}
}
