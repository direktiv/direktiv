package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/util"
)

// ReusableFunctionDefinition defines a reusable function and the fields it requires.
type ReusableFunctionDefinition struct {
	Type          FunctionType               `json:"type"           yaml:"type"`
	ID            string                     `json:"id"             yaml:"id"`
	Image         string                     `json:"image"          yaml:"image"`
	Size          Size                       `json:"size,omitempty" yaml:"size,omitempty"`
	Cmd           string                     `json:"cmd,omitempty"  yaml:"cmd,omitempty"`
	Envs          []core.EnvironmentVariable `json:"envs,omitempty" yaml:"envs,omitempty"`
	PostStartExec []string                   `json:"post_start_exec,omitempty" yaml:"post_start_exec,omitempty"`
}

// GetID returns the ID of a reusable function.
func (o *ReusableFunctionDefinition) GetID() string {
	return o.ID
}

// GetType returns the Type of function.
func (o *ReusableFunctionDefinition) GetType() FunctionType {
	return ReusableContainerFunctionType
}

// Validate validates the reusable function definition.
func (o *ReusableFunctionDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.ID == "" {
		return errors.New("id required")
	}

	if ok := util.MatchesRegex(o.ID); !ok {
		return fmt.Errorf("function id must match regex: %s", util.RegexPattern)
	}

	if o.Image == "" {
		return errors.New("image required")
	}

	return nil
}
