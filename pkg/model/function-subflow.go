package model

import (
	"errors"
	"fmt"
	"regexp"
)

type SublowFunctionDefinition struct {
	Type     FunctionType `yaml:"type"`
	ID       string       `yaml:"id"`
	Workflow string       `yaml:"workflow"`
}

func (o *SublowFunctionDefinition) GetID() string {
	return o.ID
}

func (o *SublowFunctionDefinition) GetType() FunctionType {
	return ReusableContainerFunctionType
}

func (o *SublowFunctionDefinition) Validate() error {
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

	return nil

}
