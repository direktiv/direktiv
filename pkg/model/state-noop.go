package model

import (
	"fmt"
)

type NoopState struct {
	StateCommon `yaml:",inline"`
	Transform   string            `yaml:"transform,omitempty"`
	Transition  string            `yaml:"transition,omitempty"`
	Catch       []ErrorDefinition `yaml:"catch,omitempty"`
}

func (o *NoopState) GetID() string {
	return o.ID
}

func (o *NoopState) getTransitions() map[string]string {
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

func (o *NoopState) GetTransitions() []string {
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

func (o *NoopState) ErrorDefinitions() []ErrorDefinition {
	if o.Catch == nil {
		return make([]ErrorDefinition, 0)
	}

	return o.Catch
}

func (o *NoopState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if err := validateTransformJQ(o.Transform); err != nil {
		return err
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}

	return nil
}
