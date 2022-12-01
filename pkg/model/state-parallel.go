package model

import (
	"errors"
	"fmt"
)

type ParallelState struct {
	StateCommon `yaml:",inline"`
	Actions     []ActionDefinition `yaml:"actions"`
	Mode        BranchMode         `yaml:"mode,omitempty"`
	Timeout     string             `yaml:"timeout,omitempty"`
	Transform   interface{}        `yaml:"transform,omitempty"`
	Transition  string             `yaml:"transition,omitempty"`
}

func (o *ParallelState) GetID() string {
	return o.ID
}

func (o *ParallelState) getTransitions() map[string]string {
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

func (o *ParallelState) GetTransitions() []string {
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

func (o *ParallelState) GetActions() []ActionDefinition {
	if o.Actions == nil {
		return make([]ActionDefinition, 0)
	}

	return o.Actions
}

func (o *ParallelState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if o.Actions == nil || len(o.Actions) == 0 {
		return errors.New("actions required")
	}

	for i, action := range o.GetActions() {
		if err := action.Validate(); err != nil {
			return fmt.Errorf("action[%v] is invalid: %w", i, err)
		}
	}

	if o.Timeout != "" && !isISO8601(o.Timeout) {
		return errors.New("timeout is not a ISO8601 string")
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %w", i, err)
		}
	}

	return nil
}
