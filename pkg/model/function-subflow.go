package model

import (
	"errors"
	"fmt"
	"regexp"
)

type SubflowFunctionDefinition struct {
	Type     FunctionType `yaml:"type"`
	ID       string       `yaml:"id"`
	Workflow string       `yaml:"workflow"`
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

	matched, err := regexp.MatchString(FunctionNameRegex, o.ID)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("function id must match regex: %s", FunctionNameRegex)
	}

	return nil

}
