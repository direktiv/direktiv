package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/utils"
)

// SubflowFunctionDefinition is the object to define a Subflow Function in the workflow.
type SubflowFunctionDefinition struct {
	Type     FunctionType `json:"type"     yaml:"type"`
	ID       string       `json:"id"       yaml:"id"`
	Workflow string       `json:"workflow" yaml:"workflow"`
}

// GetID returns the id of the subflow function.
func (o *SubflowFunctionDefinition) GetID() string {
	return o.ID
}

// GetType returns the type of the subflow function.
func (o *SubflowFunctionDefinition) GetType() FunctionType {
	return SubflowFunctionType
}

// Validate validates the subflow function's arguments.
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

	if ok := utils.MatchesRegex(o.ID); !ok {
		return fmt.Errorf("function id must match regex: %s", utils.RegexPattern)
	}

	return nil
}
