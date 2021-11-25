package model

import (
	"errors"
	"fmt"
)

// ForEachState defines the fields attached for a foreach
type ForEachState struct {
	StateCommon `yaml:",inline"`
	Array       interface{}       `yaml:"array"`
	Action      *ActionDefinition `yaml:"action"`
	Timeout     string            `yaml:"timeout,omitempty"`
	Transform   interface{}       `yaml:"transform,omitempty"`
	Transition  string            `yaml:"transition,omitempty"`
}

// GetID returns the ID of the state
func (o *ForEachState) GetID() string {
	return o.ID
}

func (o *ForEachState) getTransitions() map[string]string {
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

// GetTransitions returns all transitions for the state
func (o *ForEachState) GetTransitions() []string {
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

// Validate validates the arguments for a foreach state
func (o *ForEachState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	if o.Array == "" {
		return errors.New("array required")
	}

	if o.Action == nil {
		return errors.New("action required")
	}

	if o.Timeout != "" && !isISO8601(o.Timeout) {
		return errors.New("timeout is not a ISO8601 string")
	}

	return nil
}
