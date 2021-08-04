package model

import (
	"errors"
	"fmt"
	"regexp"
)

type NamespacedFunctionDefinition struct {
	Type           FunctionType             `yaml:"type"`
	ID             string                   `yaml:"id"`
	KnativeService string                   `yaml:"knative_service"`
	Files          []FunctionFileDefinition `yaml:"files,omitempty"`
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

	matched, err := regexp.MatchString(FunctionNameRegex, o.ID)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("function id must match regex: %s", FunctionNameRegex)
	}

	if o.KnativeService == "" {
		return errors.New("knative_service required")
	}

	for i, f := range o.Files {
		err := f.Validate()
		if err != nil {
			return fmt.Errorf("function file %d: %v", i, err)
		}
	}

	return nil

}
