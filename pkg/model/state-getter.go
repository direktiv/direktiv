package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/utils"
)

// GetterState defines the state for a getter.
type GetterState struct {
	StateCommon `yaml:",inline"`
	Variables   []GetterDefinition `yaml:"variables"`
	Transform   interface{}        `yaml:"transform,omitempty"`
	Transition  string             `yaml:"transition,omitempty"`
}

// GetterDefinition takes a scope and key to work out where the variable goes.
type GetterDefinition struct {
	Scope string      `yaml:"scope,omitempty"`
	Key   interface{} `yaml:"key"`
	As    string      `yaml:"as"`
}

// Validate validates against the getter definition.
func (o *GetterDefinition) Validate() error {
	switch o.Scope {
	case utils.VarScopeInstance:
	case utils.VarScopeWorkflow:
	case utils.VarScopeNamespace:
	case utils.VarScopeThread:
	case utils.VarScopeSystem:
	case utils.VarScopeFileSystem:
	default:
		return ErrVarScope
	}

	if o.Key == nil || o.Key == "" {
		return errors.New(`key required`)
	}

	return nil
}

// GetID returns the ID of the getter state.
func (o *GetterState) GetID() string {
	return o.ID
}

func (o *GetterState) getTransitions() map[string]string {
	transitions := make(map[string]string)
	if o.Transition != "" {
		transitions["transition"] = o.Transition
	}

	for i, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions[fmt.Sprintf("errors[%v]", i)] = errDef.Transition
		}
	}

	return transitions
}

// GetTransitions returns all the transitions of a getter state.
func (o *GetterState) GetTransitions() []string {
	transitions := make([]string, 0)
	if o.Transition != "" {
		transitions = append(transitions, o.Transition)
	}

	for _, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions = append(transitions, errDef.Transition)
		}
	}

	return transitions
}

// Validate validates the arguments against a getter state.
func (o *GetterState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if len(o.Variables) == 0 {
		return errors.New("variables required")
	}

	for i, varDef := range o.Variables {
		if err := varDef.Validate(); err != nil {
			return fmt.Errorf("variables[%d] is invalid: %w", i, err)
		}
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %w", i, err)
		}
	}

	return nil
}
