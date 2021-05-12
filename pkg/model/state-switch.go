package model

import (
	"errors"
	"fmt"
)

type SwitchConditionDefinition struct {
	Condition  string `yaml:"condition"`
	Transform  string `yaml:"transform,omitempty"`
	Transition string `yaml:"transition,omitempty"`
}

func (o *SwitchConditionDefinition) Validate() error {
	if o.Condition == "" {
		return errors.New("condition required")
	}

	if err := validateTransformJQ(o.Transform); err != nil {
		return err
	}

	return nil
}

type SwitchState struct {
	StateCommon       `yaml:",inline"`
	Conditions        []SwitchConditionDefinition `yaml:"conditions"`
	DefaultTransform  string                      `yaml:"defaultTransform,omitempty"`
	DefaultTransition string                      `yaml:"defaultTransition,omitempty"`
}

func (o *SwitchState) GetID() string {
	return o.ID
}

func (o *SwitchState) getTransitions() map[string]string {
	transitions := make(map[string]string)
	if o.DefaultTransition != "" {
		transitions["defaultTransition"] = o.DefaultTransition
	}

	for i, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions[fmt.Sprintf("errors[%v]", i)] = errDef.Transition
		}
	}

	for i, condition := range o.GetConditions() {
		if condition.Transition != "" {
			transitions[fmt.Sprintf("conditions[%v]", i)] = condition.Transition
		}
	}
	return transitions
}

func (o *SwitchState) GetTransitions() []string {
	transitions := make([]string, 0)
	if o.DefaultTransition != "" {
		transitions = append(transitions, o.DefaultTransition)
	}

	for _, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions = append(transitions, errDef.Transition)
		}
	}

	for _, condition := range o.GetConditions() {
		if condition.Transition != "" {
			transitions = append(transitions, condition.Transition)
		}
	}
	return transitions
}

func (o *SwitchState) GetConditions() []SwitchConditionDefinition {
	if o.Conditions == nil {
		return make([]SwitchConditionDefinition, 0)
	}

	return o.Conditions
}

func (o *SwitchState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if err := validateTransformJQ(o.DefaultTransform); err != nil {
		return fmt.Errorf("default transform: %v", err)
	}

	if o.Conditions == nil || len(o.Conditions) == 0 {
		return errors.New("conditions required")
	}

	for i, condition := range o.GetConditions() {
		if err := condition.Validate(); err != nil {
			return fmt.Errorf("conditions[%v] is invalid: %v", i, err)
		}
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}

	return nil
}
