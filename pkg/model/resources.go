package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
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
		// it is specified as openapi file, no validation here
		config := datamodel.DocumentConfiguration{
			AvoidIndexBuild:       true,
			AllowFileReferences:   true,
			AllowRemoteReferences: false,
		}

		document, err := libopenapi.NewDocumentWithConfiguration(data, &config)
		if err != nil {
			// can not fail here, so we ignoring the error
			doc, _ := libopenapi.NewDocument([]byte(fmt.Sprintf("openapi: 3.0.0\nx-direktiv-api: %s", s)))
			return doc, err
		}

		return document, nil

	case EndpointAPIV2:
		var pathItemMap map[string]interface{}
		err = yaml.Unmarshal(data, &pathItemMap)
		if err != nil {
			return &core.EndpointConfig{}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		// convert to JSON for openapi library
		b, err := json.Marshal(pathItemMap)
		if err != nil {
			return &core.EndpointConfig{}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		// source item
		var pathItem openapi3.PathItem
		err = pathItem.UnmarshalJSON(b)
		if err != nil {
			return &core.EndpointConfig{}, fmt.Errorf("error parsing direktiv resource (%s): %w", s, err)
		}

		return &pathItem, nil
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
