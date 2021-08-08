package model

import (
	"errors"
	"fmt"

	"github.com/vorteil/direktiv/pkg/util"
)

type GlobalFunctionDefinition struct {
	Type           FunctionType             `yaml:"type" json:"type"`
	ID             string                   `yaml:"id" json:"id"`
	KnativeService string                   `yaml:"service" json:"service"`
	Files          []FunctionFileDefinition `yaml:"files,omitempty" json:"files,omitempty"`
}

func (o *GlobalFunctionDefinition) GetID() string {
	return o.ID
}

func (o *GlobalFunctionDefinition) GetType() FunctionType {
	return GlobalKnativeFunctionType
}

func (o *GlobalFunctionDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.ID == "" {
		return errors.New("id required")
	}

	if ok := util.MatchesRegex(o.ID); !ok {
		return fmt.Errorf("function id must match regex: %s", util.RegexPattern)
	}

	if o.KnativeService == "" {
		return errors.New("service required")
	}

	for i, f := range o.Files {
		err := f.Validate()
		if err != nil {
			return fmt.Errorf("function file %d: %v", i, err)
		}
	}

	return nil

}
