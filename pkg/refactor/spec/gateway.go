package spec

import (
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"gopkg.in/yaml.v3"
)

type EndpointFile struct {
	core.EndpointBase
	DirektivAPI string `json:"direktiv_api,omitempty" yaml:"direktiv_api"`
}

type ConsumerFile struct {
	core.ConsumerBase
	DirektivAPI string `yaml:"direktiv_api"`
}

func ParseConsumerFile(data []byte) (*ConsumerFile, error) {
	res := &ConsumerFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "consumer/v1") {
		return nil, fmt.Errorf("invalid consumer api version")
	}

	// to avoid the ugliness of the composition struct
	err = yaml.Unmarshal(data, &res.ConsumerBase)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func ParseEndpointFile(data []byte) (*EndpointFile, error) {
	res := &EndpointFile{}
	err := yaml.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(res.DirektivAPI, "endpoint/v1") {
		return nil, fmt.Errorf("invalid endpoint api version")
	}

	// to avoid the ugliness of the composition struct
	err = yaml.Unmarshal(data, &res.EndpointBase)
	if err != nil {
		return nil, err
	}

	return res, nil
}
