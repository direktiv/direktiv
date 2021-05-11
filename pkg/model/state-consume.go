package model

import (
	"errors"
	"fmt"
)

type ConsumeEventState struct {
	StateCommon `yaml:",inline"`
	Event       *ConsumeEventDefinition `yaml:"event"`
	Timeout     string                  `yaml:"timeout,omitempty"`
	Transform   string                  `yaml:"transform,omitempty"`
	Transition  string                  `yaml:"transition,omitempty"`
}

func (o *ConsumeEventState) GetID() string {
	return o.ID
}

func (o *ConsumeEventState) getTransitions() map[string]string {
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

func (o *ConsumeEventState) GetTransitions() []string {
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

func (o *ConsumeEventState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if err := validateTransformJQ(o.Transform); err != nil {
		return err
	}

	if o.Event == nil {
		return errors.New("event required")
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}

	return nil
}
