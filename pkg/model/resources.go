package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/low"
	v3low "github.com/pb33f/libopenapi/datamodel/low/v3"
	"github.com/pb33f/libopenapi/index"
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
		// we check for libopenapi compatibility only
		_, err := libopenapi.NewDocument(data)

		// m, errs := doc.BuildV3Model()

		fmt.Println("GATEWAY!!")
		// d, _ := m.Model.Paths.PathItems.Get("/user")
		// a, _ := d.Get.Responses.Render()
		// fmt.Println(string(a))
		// fmt.Println(errs)
		// fmt.Println(m.Model.Paths.PathItems)
		return core.Gateway{}, err
	case EndpointAPIV2:
		var (
			// pathItem openapi3.PathItem
			interim  map[string]interface{}
			endpoint core.Endpoint
		)
		fmt.Println("ENDPOINT!!!!!")

		// convert yaml to json for loading
		err := yaml.Unmarshal(data, &interim)
		if err != nil {
			return &endpoint, err
		}

		// rd := index.NewRolodex(&index.SpecIndexConfig{
		// 	BasePath:          "/",
		// 	AllowRemoteLookup: true,
		// 	AvoidBuildIndex:   true,
		// 	AllowFileLookup:   true,
		// })

		// spxindex := index.SpecIndexConfig{
		// 	BasePath:          "/",
		// 	AllowRemoteLookup: true,
		// 	AvoidBuildIndex:   true,
		// 	AllowFileLookup:   true,
		// 	// Rolodex:           rd,
		// }

		// doc, _ := libopenapi.NewDocumentWithConfiguration([]byte("openapi: 3.0.0\ninfo:\n   version: \"1.0\"\n   title: dummy\npaths: {}\n"), &docConfig)
		// highDoc, _ := doc.BuildV3Model()
		// fmt.Println(highDoc)

		var idxNode yaml.Node
		err = yaml.Unmarshal(data, &idxNode)
		if err != nil {
			return &endpoint, err
		}
		// idx := index.NewSpecIndexWithConfig(&idxNode, &spxindex)
		idx := index.NewSpecIndex(&idxNode)

		var n v3low.PathItem
		err = low.BuildModel(idxNode.Content[0], &n)
		if err != nil {
			return &endpoint, err
		}

		// idx.GetRolodex().AddLocalFS("/", &gateway.DirektivOpenAPIFS{})

		fmt.Printf("ROLODEX %v\n", idx.GetRolodex())

		fmt.Printf("BUILD ERR %v\n", err)
		// err = n.Build(context.Background(), nil, idxNode.Content[0], idx)
		// fmt.Printf("BUILD ERR %v\n", err)

		// fmt.Printf("AAAA1 %+v\n", n.Get.Value.Responses)
		// bb := n.Get.Value.Responses.Value.Codes.OrderedMap.Len()
		// fmt.Printf("AAAA1 %+v\n", bb)
		// fmt.Printf("AAAA1 %+v\n", n.Get.Value.Responses.ValueNode)
		// fmt.Printf("AAAA2 %+v\n", n.Get.ValueNode)

		// pi := v3high.NewPathItem(&n)
		// gg, _ := pi.MarshalYAML()
		// out, _ := yaml.Marshal(gg)
		// fmt.Printf("YAML %+v\n", string(out))

		// fmt.Printf("FFF %+v\n", pi.Ma)

		// fmt.Println(highDoc)
		// highDoc.Model.Paths.PathItems = &orderedmap.Map[string, *v3high.PathItem]{}
		// pathItem := v3high.NewPathItem(&n)
		// highDoc.Model.Paths.PathItems.Set("/dummy", pathItem)

		// a, _ := pathItem.Render()
		// fmt.Println(string(a))

		return endpoint, nil
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
