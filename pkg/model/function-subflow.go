package model

import (
	"errors"
	"fmt"

	"github.com/vorteil/direktiv/pkg/util"
)

type SubflowFunctionDefinition struct {
	Type     FunctionType `yaml:"type" json:"type"`
	ID       string       `yaml:"id" json:"id"`
	Workflow string       `yaml:"workflow" json:"workflow"`
}

func (o *SubflowFunctionDefinition) GetID() string {
	return o.ID
}

func (o *SubflowFunctionDefinition) GetType() FunctionType {
	return SubflowFunctionType
}

func (o *SubflowFunctionDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.ID == "" {
		return errors.New("id required")
	}

	if o.Workflow == "" {
		return errors.New("workflow required")
	}

	if ok := util.MatchesRegex(o.ID); !ok {
		return fmt.Errorf("function id must match regex: %s", util.RegexPattern)
	}

	return nil

}
