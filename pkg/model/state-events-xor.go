package model

import (
	"errors"
	"fmt"
)

type EventConditionDefinition struct {
	Event      ConsumeEventDefinition `yaml:"event"`
	Transform  string                 `yaml:"transform,omitempty"`
	Transition string                 `yaml:"transition,omitempty"`
}

type EventsXorState struct {
	StateCommon `yaml:",inline"`
	Events      []EventConditionDefinition `yaml:"events"`
	Timeout     string                     `yaml:"timeout,omitempty"`
	Catch       []ErrorDefinition          `yaml:"catch,omitempty"`
}

func (o *EventsXorState) GetID() string {
	return o.ID
}

func (o *EventsXorState) getTransitions() map[string]string {
	transitions := make(map[string]string)

	for i, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions[fmt.Sprintf("errors[%v]", i)] = errDef.Transition
		}
	}

	for i, event := range o.GetEvents() {
		if event.Transition != "" {
			transitions[fmt.Sprintf("events[%v]", i)] = event.Transition
		}
	}

	return transitions
}

func (o *EventsXorState) GetTransitions() []string {
	transitions := make([]string, 0)

	for _, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions = append(transitions, errDef.Transition)
		}
	}

	for _, event := range o.GetEvents() {
		if event.Transition != "" {
			transitions = append(transitions, event.Transition)
		}
	}

	return transitions
}

func (o *EventsXorState) ErrorDefinitions() []ErrorDefinition {
	if o.Catch == nil {
		return make([]ErrorDefinition, 0)
	}

	return o.Catch
}

func (o *EventsXorState) GetEvents() []EventConditionDefinition {
	if o.Events == nil {
		return make([]EventConditionDefinition, 0)
	}

	return o.Events
}

func (o *EventsXorState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if o.Timeout != "" && !isISO8601(o.Timeout) {
		return errors.New("timeout is not a ISO8601 string")
	}

	if len(o.GetEvents()) == 0 {
		return errors.New("atleast one event is required")
	}

	for i, event := range o.GetEvents() {
		if err := validateTransformJQ(event.Transform); err != nil {
			return fmt.Errorf("event[%v]: %v", i, err)
		}

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
