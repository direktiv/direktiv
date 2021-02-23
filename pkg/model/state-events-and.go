package model

import (
	"errors"
	"fmt"
)

type EventsAndState struct {
	StateCommon `yaml:",inline"`
	Events      []EventConditionDefinition `yaml:"events"`
	Timeout     string                     `yaml:"timeout,omitempty"`
	Transform   string                     `yaml:"transform,omitempty"`
	Transition  string                     `yaml:"transition,omitempty"`
	Catch       []ErrorDefinition          `yaml:"catch,omitempty"`
}

func (o *EventsAndState) GetID() string {
	return o.ID
}

func (o *EventsAndState) getTransitions() map[string]string {
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

func (o *EventsAndState) GetTransitions() []string {
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

func (o *EventsAndState) ErrorDefinitions() []ErrorDefinition {
	if o.Catch == nil {
		return make([]ErrorDefinition, 0)
	}

	return o.Catch
}

func (o *EventsAndState) GetEvents() []EventConditionDefinition {
	if o.Events == nil {
		return make([]EventConditionDefinition, 0)
	}

	return o.Events
}

func (o *EventsAndState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if err := validateTransformJQ(o.Transform); err != nil {
		return err
	}

	if o.Timeout != "" && !isISO8601(o.Timeout) {
		return errors.New("timeout is not a ISO8601 string")
	}

	if len(o.GetEvents()) == 0 {
		return errors.New("atleast one event is required")
	}

	for i, event := range o.GetEvents() {
		if err := event.Event.Validate(); err != nil {
			return fmt.Errorf("event[%v] is invalid: %v", i, err)
		}
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}

	return nil
}
