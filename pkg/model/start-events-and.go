package model

import (
	"errors"
)

type EventsAndStart struct {
	StartCommon `yaml:",inline"`
	LifeSpan    string                 `yaml:"lifespan,omitempty"`
	Events      []StartEventDefinition `yaml:"events"`
}

func (o *EventsAndStart) GetEvents() []StartEventDefinition {
	events := make([]StartEventDefinition, 0)
	if o != nil && o.Events != nil {
		for i := range o.Events {
			events = append(events, o.Events[i])
		}
	}
	return events
}

func (o *EventsAndStart) Validate() error {
	if o.Events == nil || len(o.Events) == 0 {
		return errors.New("events required")
	}

	if o.LifeSpan != "" && !isISO8601(o.LifeSpan) {
		return errors.New("lifespan is not a ISO8601 string")
	}

	if err := o.commonValidate(); err != nil {
		return err
	}

	return nil
}
