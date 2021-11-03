package model

import (
	"errors"
	"fmt"
)

// ValidateState...
type ValidateState struct {
	StateCommon `yaml:",inline"`
	Subject     string      `yaml:"subject,omitempty"`
	Schema      interface{} `yaml:"schema"`
	Transform   interface{} `yaml:"transform,omitempty"`
	Transition  string      `yaml:"transition,omitempty"`
}

// GetID...
func (o *ValidateState) GetID() string {
	return o.ID
}

func (o *ValidateState) getTransitions() map[string]string {
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

// GetTransitions...
func (o *ValidateState) GetTransitions() []string {
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

// Validate...
func (o *ValidateState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	if o.Schema == nil {
		return errors.New("schema required")
	}

	if err := isJSONSchema(o.Schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}

	return nil
}
