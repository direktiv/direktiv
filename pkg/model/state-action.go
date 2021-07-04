package model

import (
	"fmt"

	"errors"
)

type ActionState struct {
	StateCommon `yaml:",inline"`
	Action      *ActionDefinition `yaml:"action"`
	Async       bool              `yaml:"async"`
	Timeout     string            `yaml:"timeout,omitempty"`
	Transform   interface{}       `yaml:"transform,omitempty"`
	Transition  string            `yaml:"transition,omitempty"`
}

func (o *ActionState) GetID() string {
	return o.ID
}

func (o *ActionState) getTransitions() map[string]string {
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

func (o *ActionState) GetTransitions() []string {
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

func (o *ActionState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if o.Action == nil {
		return errors.New("action required")
	}

	if o.Timeout != "" && !isISO8601(o.Timeout) {
		return errors.New("timeout is not a ISO8601 string")
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}

	return nil
}
