package model

import (
	"errors"
	"fmt"
	"regexp"
)

type GlobalFunctionDefinition struct {
	Type           FunctionType             `yaml:"type"`
	ID             string                   `yaml:"id"`
	KnativeService string                   `yaml:"knative_service"`
	Files          []FunctionFileDefinition `yaml:"files,omitempty"`
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
