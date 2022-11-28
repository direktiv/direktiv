package model

import (
	"errors"
	"fmt"
)

type EventsAndState struct {
	StateCommon `yaml:",inline"`
	Events      []ConsumeEventDefinition `yaml:"events"`
	Timeout     string                   `yaml:"timeout,omitempty"`
	Transform   interface{}              `yaml:"transform,omitempty"`
	Transition  string                   `yaml:"transition,omitempty"`
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

func (o *EventsAndState) GetEvents() []ConsumeEventDefinition {
	if o.Events == nil {
		return make([]ConsumeEventDefinition, 0)
	}

	return o.Events
}

func (o *EventsAndState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	if o.Timeout != "" && !isISO8601(o.Timeout) {
		return errors.New("timeout is not a ISO8601 string")
	}

	if len(o.GetEvents()) == 0 {
		return errors.New("at least one event is required")
	}

	for i, event := range o.GetEvents() {
		if err := event.Validate(); err != nil {
			return fmt.Errorf("event[%v] is invalid: %w", i, err)
		}
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %w", i, err)
		}
	}

	return nil
}
