package model

import (
	"errors"
	"fmt"
)

type GenerateEventDefinition struct {
	Type            string                 `yaml:"type"`
	Source          string                 `yaml:"source"`
	Data            interface{}            `yaml:"data,omitempty"`
	DataContentType string                 `yaml:"data_content_type,omitempty"`
	Context         map[string]interface{} `yaml:"context,omitempty"`
}

func (o *GenerateEventDefinition) Validate() error {
	if o.Type == "" {
		return errors.New("type required")
	}

	if o.Source == "" {
		return errors.New("source required")
	}

	return nil
}

type GenerateEventState struct {
	StateCommon `yaml:",inline"`
	Delay       string                   `yaml:"delay"`
	Event       *GenerateEventDefinition `yaml:"event"`
	Transform   interface{}              `yaml:"transform,omitempty"`
	Transition  string                   `yaml:"transition,omitempty"`
}

func (o *GenerateEventState) GetID() string {
	return o.ID
}

func (o *GenerateEventState) GetDelay() string {
	return o.Delay
}

func (o *GenerateEventState) getTransitions() map[string]string {
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

func (o *GenerateEventState) GetTransitions() []string {
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

func (o *GenerateEventState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	if o.Delay != "immediate" && o.Delay != "" && !isISO8601(o.Delay) {
		return errors.New("delay is not a ISO8601 string")
	}

	if o.Event == nil {
		return errors.New("event required")
	}

	err := o.Event.Validate()
	if err != nil {
		return err
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}
	return nil
}
