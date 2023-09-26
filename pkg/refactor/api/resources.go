package api

import (
	"errors"
	"fmt"
	"strings"

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
	WorkflowAPIV1 = "workflow/v1"
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
		wf := new(model.Workflow)
		err = wf.Load(data)
		if err != nil {
			return &model.Workflow{
				DirektivAPI: s,
			}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return wf, nil
	default:
		return nil, fmt.Errorf("error parsing direktiv resource: invalid 'direktiv_api': \"%s\"", s)
	}
}

type Service struct {
	DirektivAPI string `yaml:"direktiv_api"`
	Name        string `yaml:"name"`
	Image       string `yaml:"image"`
	Scale       int32  `yaml:"scale"`
	Size        string `yaml:"size"`
	Cmd         string `yaml:"cmd"`
}

func ParseService(data []byte) *Service {
	res := &Service{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil
	}
	if strings.HasPrefix(res.DirektivAPI, "service/") {
		return nil
	}

	return res
}
