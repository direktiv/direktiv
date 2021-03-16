package model

import (
	"errors"
	"fmt"

	"github.com/itchyny/gojq"
)

type RetryDefinition struct {
	MaxAttempts int     `yaml:"max_attempts"`
	Delay       string  `yaml:"delay,omitempty"`
	Multiplier  float64 `yaml:"multiplier,omitempty"`
}

func (o *RetryDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.MaxAttempts == 0 {
		return errors.New("maxAttempts required to be more than 0")
	}

	if o.Delay != "" && !isISO8601(o.Delay) {
		return errors.New("delay is not a ISO8601 string")
	}

	return nil
}

type ErrorDefinition struct {
	Error      string           `yaml:"error"`
	Retry      *RetryDefinition `yaml:"retry,omitempty"`
	Transition string           `yaml:"transition,omitempty"`
}

func (o *ErrorDefinition) Validate() error {
	if o.Error == "" {
		return errors.New("error required")
	}

	if err := o.Retry.Validate(); err != nil {
		return err
	}

	return nil
}

type State interface {
	GetID() string
	GetType() StateType
	Validate() error
	ErrorDefinitions() []ErrorDefinition
	GetTransitions() []string
	getTransitions() map[string]string
}

type ConsumeEventDefinition struct {
	Type    string                 `yaml:"type"`
	Context map[string]interface{} `yaml:"context,omitempty"`
}

func (o *ConsumeEventDefinition) Validate() error {
	if o.Type == "" {
		return errors.New("type required")
	}

	return nil

}

type ProduceEventDefinition struct {
	Type    string                 `yaml:"type,omitempty"`
	Source  string                 `yaml:"source,omitempty"`
	Data    string                 `yaml:"data"`
	Context map[string]interface{} `yaml:"context"`
}

func (o *ProduceEventDefinition) Validate() error {
	if o.Source == "" {
		return errors.New("source required")
	}

	if o.Type == "" {
		return errors.New("type required")
	}

	return nil

}

type StateCommon struct {
	ID   string    `yaml:"id"`
	Type StateType `yaml:"type"`
	Log  string    `yaml:"log"`
}

func (o *StateCommon) GetType() StateType {
	return o.Type
}

func (o *StateCommon) commonValidate() error {
	if o.ID == "" {
		return errors.New("id required")
	}

	if o.Log != "" {
		if _, err := gojq.Parse(o.Log); err != nil {
			return fmt.Errorf("log is an invalid jq string: %v", err)
		}
	}

	return nil
}

// util
func getStateFromType(stype string) (State, error) {
	var s State
	var err error

	switch stype {
	case StateTypeSwitch.String():
		s = new(SwitchState)
	case StateTypeForEach.String():
		s = new(ForEachState)
	case StateTypeAction.String():
		s = new(ActionState)
	case StateTypeConsume.String():
		s = new(ConsumeEventState)
	case StateTypeDelay.String():
		s = new(DelayState)
	case StateTypeEventsAnd.String():
		s = new(EventsAndState)
	case StateTypeEventsXor.String():
		s = new(EventsXorState)
	case StateTypeError.String():
		s = new(ErrorState)
	case StateTypeGenerateEvent.String():
		s = new(GenerateEventState)
	case StateTypeNoop.String():
		s = new(NoopState)
	case StateTypeValidate.String():
		s = new(ValidateState)
	case StateTypeCallback.String():
		s = new(CallbackState)
	case StateTypeParallel.String():
		s = new(ParallelState)
	case "":
		err = errors.New("type required")
	default:
		err = errors.New("type unimplemented/unrecognized")
	}

	return s, err
}
