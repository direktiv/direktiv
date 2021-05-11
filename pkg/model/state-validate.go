package model

import (
	"errors"
	"fmt"
)

type ValidateState struct {
	StateCommon `yaml:",inline"`
	Subject     string      `yaml:"subject"`
	Schema      interface{} `yaml:"schema"`
	Transform   string      `yaml:"transform,omitempty"`
	Transition  string      `yaml:"transition,omitempty"`
}

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

func (o *ValidateState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if err := validateTransformJQ(o.Transform); err != nil {
		return err
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
