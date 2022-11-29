package model

import (
	"errors"
	"fmt"
)

type DelayState struct {
	StateCommon `yaml:",inline"`
	Duration    string      `yaml:"duration"`
	Transform   interface{} `yaml:"transform,omitempty"`
	Transition  string      `yaml:"transition,omitempty"`
}

func (o *DelayState) GetID() string {
	return o.ID
}

func (o *DelayState) getTransitions() map[string]string {
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

func (o *DelayState) GetTransitions() []string {
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

func (o *DelayState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if o.Duration == "" {
		return errors.New("duration required")
	}

	if !isISO8601(o.Duration) {
		return errors.New("duration is not a ISO8601 string")
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %w", i, err)
		}
	}

	return nil
}
