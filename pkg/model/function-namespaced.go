package model

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/direktiv/direktiv/pkg/utils"
)

// NamespacedFunctionDefinition defines a namespace service in the workflow.
type NamespacedFunctionDefinition struct {
	Type FunctionType `json:"type"    yaml:"type"`
	ID   string       `json:"id"      yaml:"id"`
	Path string       `json:"service" yaml:"service"` // NOTE: 'service' instead of 'path' for minor compatibility reasons. Change this later, perhaps.
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

	if ok := utils.MatchesRegex(o.ID); !ok {
		return fmt.Errorf("function id must match regex: %s", utils.RegexPattern)
	}

	filePathPattern := `^/([^/]+/?)+[^/]*$`
	regex := regexp.MustCompile(filePathPattern)
	if !regex.MatchString(o.Path) {
		return fmt.Errorf("'%s' is a valid service file path", o.Path)
	}

	return nil
}
