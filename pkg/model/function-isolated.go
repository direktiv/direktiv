package model

import (
	"errors"
	"fmt"
	"regexp"
)

type IsolatedFunctionDefinition struct {
	Type  FunctionType             `yaml:"type"`
	ID    string                   `yaml:"id"`
	Image string                   `yaml:"image"`
	Size  Size                     `yaml:"size,omitempty"`
	Cmd   string                   `yaml:"cmd,omitempty"`
	Files []FunctionFileDefinition `yaml:"files,omitempty"`
}

func (o *IsolatedFunctionDefinition) GetID() string {
	return o.ID
}

func (o *IsolatedFunctionDefinition) GetType() FunctionType {
	return ReusableContainerFunctionType
}

func (o *IsolatedFunctionDefinition) Validate() error {
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

	if o.Image == "" {
		return errors.New("image required")
	}

	for i, f := range o.Files {
		err := f.Validate()
		if err != nil {
			return fmt.Errorf("function file %d: %v", i, err)
		}
	}

	return nil

}
