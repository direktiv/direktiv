package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/getkin/kin-openapi/openapi3"
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

const (
	GatewayAPIV1 = "gateway/v1"
)

const (
	EndpointAPIV2 = "endpoint/v2"
)

var ErrNotDirektivAPIResource = errors.New("not a direktiv_api resource")

func LoadResource(data []byte) (interface{}, error) {
	s, err := extractType(data)
	if err != nil {
		return nil, err
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
		return &core.EndpointConfig{}, fmt.Errorf("envpoint/v1 not supported anymore (%s)", s)
	case ConsumerAPIV1:
		ef := new(core.ConsumerFile)
		err = yaml.Unmarshal(data, &ef)
		if err != nil {
			return &core.ConsumerFile{
				DirektivAPI: s,
			}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return ef, nil

	case GatewayAPIV1:
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
			return []byte(""), nil
		}
		_, err = loader.LoadFromData(data)

		return core.Gateway{}, err

	case EndpointAPIV2:
		var pi openapi3.PathItem

		// its yaml but we need JSON
		var interim map[string]interface{}
		err := yaml.Unmarshal(data, &interim)
		if err != nil {
			return core.Endpoint{}, err
		}
		d, err := json.Marshal(interim)
		if err != nil {
			return core.Endpoint{}, err
		}

		err = pi.UnmarshalJSON(d)
		if err != nil {
			return core.Endpoint{}, err
		}

		return core.Endpoint{}, err

	default:
		return nil, fmt.Errorf("error parsing direktiv resource: invalid 'direktiv_api': \"%s\"", s)
	}
}

func extractType(data []byte) (string, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNotDirektivAPIResource, err)
	}

	// check for openapi gateway resource or regular resource
	x, exists := m["direktiv_api"]
	if !exists {
		x, exists = m["x-direktiv-api"]
	}

	if !exists {
		return "", fmt.Errorf("%w: missing 'direktiv_api' field",
			ErrNotDirektivAPIResource)
	}

	s, ok := x.(string)
	if !ok {
		return "", fmt.Errorf("%w: invalid 'direktiv_api' field",
			ErrNotDirektivAPIResource)
	}

	return s, nil
}
