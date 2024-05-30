package model

import (
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/utils"
)

// RetryDefinition defines a retry object to be used in the workflow.
type RetryDefinition struct {
	MaxAttempts int      `json:"max_attempts" yaml:"max_attempts"`
	Delay       string   `json:"delay"        yaml:"delay,omitempty"`
	Multiplier  float64  `json:"multiplier"   yaml:"multiplier,omitempty"`
	Codes       []string `json:"codes"        yaml:"codes"`
}

// Validate checks the arguments for the retry definition.
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

	if len(o.Codes) == 0 {
		return errors.New("retry policy requires at least one defined code")
	}

	return nil
}

// ErrorDefinition defines an error object to be used in the workflow.
type ErrorDefinition struct {
	Error      string `yaml:"error"`
	Transition string `yaml:"transition,omitempty"`
}

// Validate checks the arguments for the error definition.
func (o *ErrorDefinition) Validate() error {
	if o.Error == "" {
		return errors.New("error required")
	}

	return nil
}

// State a simple interface to define a state.
type State interface {
	GetID() string
	GetType() StateType
	Validate() error
	ErrorDefinitions() []ErrorDefinition
	GetTransitions() []string
	getTransitions() map[string]string
}

// ConsumeEventDefinition defines what a consume event is.
type ConsumeEventDefinition struct {
	Type    string                 `yaml:"type"`
	Context map[string]interface{} `yaml:"context,omitempty"`
}

// Validate validates the arguments provided for the consume event definition.
func (o *ConsumeEventDefinition) Validate() error {
	if o.Type == "" {
		return errors.New("type required")
	}

	return nil
}

// ProduceEventDefinition defines what a produce event is.
type ProduceEventDefinition struct {
	Type    string                 `yaml:"type,omitempty"`
	Source  string                 `yaml:"source,omitempty"`
	Data    string                 `yaml:"data,omitempty"`
	Context map[string]interface{} `yaml:"context,omitempty"`
}

// Validate validates the arguments provided for the produce event definition.
func (o *ProduceEventDefinition) Validate() error {
	if o.Source == "" {
		return errors.New("source required")
	}

	if o.Type == "" {
		return errors.New("type required")
	}

	return nil
}

// StateCommon defines the common attributes of a state.
type StateCommon struct {
	ID       string            `yaml:"id"`
	Type     StateType         `yaml:"type"`
	Log      interface{}       `yaml:"log,omitempty"`
	Metadata interface{}       `yaml:"metadata,omitempty"`
	Catch    []ErrorDefinition `yaml:"catch,omitempty"`
}

// GetType returns the type of a state common.
func (o *StateCommon) GetType() StateType {
	return o.Type
}

// GetLog returns the log query.
func (o *StateCommon) GetLog() interface{} {
	return o.Log
}

// GetMetadata returns the metadata query.
func (o *StateCommon) GetMetadata() interface{} {
	return o.Metadata
}

// ErrorDefinitions returns an array of error definitions.
func (o *StateCommon) ErrorDefinitions() []ErrorDefinition {
	if o.Catch == nil {
		return make([]ErrorDefinition, 0)
	}

	return o.Catch
}

func (o *StateCommon) commonValidate() error {
	if o.ID == "" {
		return errors.New("id required")
	}

	if ok := utils.MatchesRegex(o.ID); !ok {
		return fmt.Errorf("state id must match the regex pattern `%s`", utils.RegexPattern)
	}

	for _, catch := range o.Catch {
		if err := catch.Validate(); err != nil {
			return err
		}
	}

	return nil
}

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
	case "eventAnd":
		fallthrough
	case StateTypeEventsAnd.String():
		s = new(EventsAndState)
	case "eventXor":
		fallthrough
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
	case StateTypeParallel.String():
		s = new(ParallelState)
	case StateTypeGetter.String():
		s = new(GetterState)
	case StateTypeSetter.String():
		s = new(SetterState)
	case "":
		err = errors.New("type required(switch, foreach, consumeEvent, delay, eventsAnd, eventsXor, noop, validate, parallel, getter, setter)")
	default:
		err = errors.New("type unrecognized(switch, foreach, consumeEvent, delay, eventsAnd, eventsXor, noop, validate, parallel, getter, setter)")
	}

	return s, err
}
