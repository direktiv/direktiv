package model

import "errors"

type EventsXorStart struct {
	StartCommon `yaml:",inline"`
	Events      []StartEventDefinition `yaml:"events"`
}

func (o *EventsXorStart) GetEvents() []StartEventDefinition {
	events := make([]StartEventDefinition, 0)
	if o != nil && o.Events != nil {
		events = append(events, o.Events...)
	}
	return events
}

func (o *EventsXorStart) Validate() error {
	if o.Events == nil || len(o.Events) == 0 {
		return errors.New("events required")
	}

	if err := o.commonValidate(); err != nil {
		return err
	}

	return nil
}
