package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/util"
)

type NamespacedFunctionDefinition struct {
	Type           FunctionType             `yaml:"type" json:"type"`
	ID             string                   `yaml:"id" json:"id"`
	KnativeService string                   `yaml:"service" json:"service"`
	Files          []FunctionFileDefinition `yaml:"files,omitempty" json:"files,omitempty"`
}

func (o *NamespacedFunctionDefinition) GetID() string {
	return o.ID
}

func (o *NamespacedFunctionDefinition) GetType() FunctionType {
	return NamespacedKnativeFunctionType
}

func (o *NamespacedFunctionDefinition) Validate() error {
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
