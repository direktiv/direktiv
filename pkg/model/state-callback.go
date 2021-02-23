package model

import (
	"errors"
	"fmt"
)

type CallbackState struct {
	StateCommon `yaml:",inline"`
	Action      *ActionDefinition       `yaml:"action"`
	Event       *ConsumeEventDefinition `yaml:"event"`
	Timeout     string                  `yaml:"timeout,omitempty"`
	Transform   string                  `yaml:"transform,omitempty"`
	Transition  string                  `yaml:"transition,omitempty"`
	Catch       []ErrorDefinition       `yaml:"catch,omitempty"`
}

func (o *CallbackState) GetID() string {
	return o.ID
}

func (o *CallbackState) getTransitions() map[string]string {
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

func (o *CallbackState) GetTransitions() []string {
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

func (o *CallbackState) ErrorDefinitions() []ErrorDefinition {
	if o.Catch == nil {
		return make([]ErrorDefinition, 0)
	}

	return o.Catch
}

func (o *CallbackState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if err := validateTransformJQ(o.Transform); err != nil {
		return err
	}

	if o.Action == nil {
		return errors.New("action required")
	}

	if o.Event == nil {
		return errors.New("event required")
	}

	if o.Timeout != "" && !isISO8601(o.Timeout) {
		return errors.New("timeout is not a ISO8601 string")
	}

	return nil
}
