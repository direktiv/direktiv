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
	Transform   interface{}             `yaml:"transform,omitempty"`
	Transition  string                  `yaml:"transition,omitempty"`
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

func (o *CallbackState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
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
