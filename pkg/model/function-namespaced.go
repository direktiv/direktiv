package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/util"
)

// NamespacedFunctionDefinition defines a namespace service in the workflow.
type NamespacedFunctionDefinition struct {
	Type           FunctionType `json:"type"    yaml:"type"`
	ID             string       `json:"id"      yaml:"id"`
	KnativeService string       `json:"service" yaml:"service"`
}

// GetID returns the id of a namespace function.
func (o *NamespacedFunctionDefinition) GetID() string {
	return o.ID
}

// GetType returns the type of the function.
func (o *NamespacedFunctionDefinition) GetType() FunctionType {
	return NamespacedKnativeFunctionType
}

// Validate validates the namespace function definition's arguments.
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

	return nil
}
