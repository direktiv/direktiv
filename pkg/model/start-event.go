package model

import "errors"

type EventStart struct {
	StartCommon `yaml:",inline"`
	Event       *StartEventDefinition `yaml:"event"`
}

func (o *EventStart) GetEvents() []StartEventDefinition {
	events := make([]StartEventDefinition, 0)
	if o != nil && o.Event != nil {
		events = append(events, *o.Event)
	}
	return events
}

func (o *EventStart) Validate() error {
	if o.Event == nil {
		return errors.New("event required")
	}

	if err := o.commonValidate(); err != nil {
		return err
	}

	return nil
}
