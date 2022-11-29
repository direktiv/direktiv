package model

import (
	"errors"
)

type StartDefinition interface {
	GetState() string
	GetType() StartType
	Validate() error
	GetEvents() []StartEventDefinition
}

func (o *Workflow) GetStartDefinition() StartDefinition {

	if o.Start != nil {
		return o.Start
	}

	return &DefaultStart{}

}

type StartEventDefinition struct {
	Type    string                 `yaml:"type"`
	Context map[string]interface{} `yaml:"context,omitempty"`
}

func (o *StartEventDefinition) Validate() error {
	if o.Type == "" {
		return errors.New("type required")
	}

	return nil
}

type StartCommon struct {
	Type  StartType `yaml:"type"`
	State string    `yaml:"state,omitempty"`
}

func (o *StartCommon) commonValidate() error {
	// if o.Type == "" {
	// 	return errors.New("type required")
	// }
	return nil
}

func (o *StartCommon) GetType() StartType {

	if o == nil {
		return StartTypeDefault
	}

	return o.Type

}

func (o *StartCommon) GetState() string {

	if o == nil {
		return ""
	}

	return o.State

}

func getStartFromType(startType string) (StartDefinition, error) {
	var s StartDefinition
	var err error

	switch startType {
	case StartTypeScheduled.String():
		s = new(ScheduledStart)
	case StartTypeEvent.String():
		s = new(EventStart)
	case StartTypeEventsXor.String():
		s = new(EventsXorStart)
	case StartTypeEventsAnd.String():
		s = new(EventsAndStart)
	case StartTypeDefault.String():
		s = new(DefaultStart)
	case "":
		err = errors.New("type required(scheduled, event, eventsXor, eventsAnd, default)")
	default:
		err = errors.New("type unrecognized(scheduled, event, eventsXor, eventsAnd, default)")
	}

	return s, err
}
