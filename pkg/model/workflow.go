package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Workflow struct {
	ID          string               `yaml:"id"`
	Name        string               `yaml:"name,omitempty"`
	Description string               `yaml:"description,omitempty"`
	Functions   []FunctionDefinition `yaml:"functions,omitempty"`
	Schemas     []SchemaDefinition   `yaml:"schemas,omitempty"`
	States      []State              `yaml:"states,omitempty"`
	Timeouts    *TimeoutDefinition   `yaml:"timeouts,omitempty"`
	Start       StartDefinition      `yaml:"start,omitempty"`
}

func (o *Workflow) unmarshal(m map[string]interface{}) error {

	// split start out from the rest
	y, startFound := m["start"]
	if startFound {
		// Start

		delete(m, "start")
		strMap, ok := y.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid start")
		}

		strType, ok := strMap["type"]
		if !ok {
			return fmt.Errorf("missing 'type' for start")
		}

		strTypeString, ok := strType.(string)
		if !ok {
			return fmt.Errorf("start bad data-format for 'type'")
		}

		strData, err := json.Marshal(strMap)
		if err != nil {
			panic(err)
		}

		var start StartDefinition

		switch strTypeString {
		case StartTypeScheduled.String():
			start = new(ScheduledStart)
		case StartTypeEvent.String():
			start = new(EventStart)
		case StartTypeEventsXor.String():
			start = new(EventsXorStart)
		case StartTypeEventsAnd.String():
			start = new(EventsAndStart)
		case "":
			return fmt.Errorf("start: type required")
		default:
			return fmt.Errorf("start: type unimplemented/unrecognized")
		}

		err = json.Unmarshal(strData, start)
		if err != nil {
			return err
		}

		err = start.Validate()
		if err != nil {
			return fmt.Errorf("start invalid: %w", err)
		}

		o.Start = start
	}

	// split states out from the rest
	x, ok := m["states"]
	if !ok {
		return errors.New("states required")
	}

	delete(m, "states")

	data, err := json.Marshal(&m)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &o)
	if err != nil {
		return err
	}

	// cast all states
	list, ok := x.([]interface{})
	if !ok {
		return errors.New("invalid type for states")
	}

	o.States = make([]State, len(list))

	for i := range list {

		sm, ok := list[i].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid state[%d]", i)
		}

		st, ok := sm["type"]
		if !ok {
			return fmt.Errorf("missing 'type' for state[%d]", i)
		}

		stype, ok := st.(string)
		if !ok {
			return fmt.Errorf("state[%d]: bad data-format for 'type'", i)
		}

		sdata, err := json.Marshal(sm)
		if err != nil {
			panic(err)
		}

		var s State

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
			return fmt.Errorf("state[%d]: type required", i)
		default:
			return fmt.Errorf("state[%d]: type unimplemented/unrecognized", i)
		}

		err = json.Unmarshal(sdata, s)
		if err != nil {
			return err
		}

		o.States[i] = s

		err = s.Validate()
		if err != nil {
			return fmt.Errorf("state[%d]: %w", i, err)
		}

	}

	err = o.validate()
	if err != nil {
		return err
	}

	return nil

}

func (o *Workflow) validate() error {
	if o.ID == "" {
		return fmt.Errorf("workflow id required")
	}

	regex := "^[a-z][a-z0-9._-]{1,34}[a-z0-9]$"

	matched, err := regexp.MatchString(regex, o.ID)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("workflow ID must match regex: %s", regex)
	}

	states, err := o.getStatesMap()
	if err != nil {
		return err
	}

	functions, err := o.getFunctionMap()
	if err != nil {
		return err
	}

	if o.Start != nil && o.Start.GetState() != "" {
		// Check if state exists
		if _, ok := states[o.Start.GetState()]; !ok {
			return fmt.Errorf("start targets state that does not exist")
		}
	}

	// functions
	for i, function := range o.GetFunctions() {
		if sErr := function.Validate(); sErr != nil {
			return fmt.Errorf("workflow function[%v] is invalid: %v", i, sErr)
		}
	}

	// schemas
	for i, schema := range o.GetSchemas() {
		if sErr := schema.Validate(); sErr != nil {
			return fmt.Errorf("workflow schema[%v] is invalid: %v", i, sErr)
		}
	}

	// states
	for i, state := range o.GetStates() {
		// Validate All State Transitions reference a exisiting state
		for tKey, transition := range state.getTransitions() {
			if _, ok := states[transition]; !ok {
				return fmt.Errorf("workflow state[%v] '%v' transition '%s' does not exist", i, tKey, transition)
			}
		}

		// Check if function actions are defined
		fActions := make([]string, 0)
		switch state.GetType() {
		case StateTypeAction:
			fActions = append(fActions, state.(*ActionState).Action.Function)
		case StateTypeParallel:
			for _, act := range state.(*ParallelState).Actions {
				fActions = append(fActions, act.Function)
			}
		case StateTypeForEach:
			fActions = append(fActions, state.(*ForEachState).Action.Function)
		}

		for j := range fActions {
			if _, fExists := functions[fActions[j]]; fActions[j] != "" && !fExists {
				return fmt.Errorf("workflow state[%v] actions function '%s' does not exist", i, fActions[j])
			}
		}

	}

	// timeout
	if sErr := o.Timeouts.Validate(); sErr != nil {
		return sErr
	}

	return nil

}

func (o *Workflow) GetStates() []State {
	if o.States == nil {
		return make([]State, 0)
	}

	return o.States
}

// GetStatesMap : Get workflow states as a map
func (o *Workflow) GetStatesMap() map[string]State {
	statesMap := make(map[string]State)
	for _, state := range o.GetStates() {
		statesMap[state.GetID()] = state
	}

	return statesMap
}

// getStatesMap : Get workflow states as a map, and returns error if the same state is defined more than once
func (o *Workflow) getStatesMap() (map[string]State, error) {
	statesMap := make(map[string]State)

	for _, state := range o.GetStates() {
		sID := state.GetID()
		if _, ok := statesMap[sID]; ok {
			return statesMap, fmt.Errorf("state id '%s' is used in more than one state", sID)
		}
		statesMap[state.GetID()] = state
	}

	return statesMap, nil
}

// getFunctionMap : Get functions as a map, and returns error if the same function id is defined more than once
func (o *Workflow) getFunctionMap() (map[string]FunctionDefinition, error) {
	funcMap := make(map[string]FunctionDefinition)

	for _, wfFunc := range o.GetFunctions() {
		fID := wfFunc.ID
		if _, ok := funcMap[fID]; ok {
			return funcMap, fmt.Errorf("function id '%s' is used in more than one function", fID)
		}
		funcMap[fID] = wfFunc
	}

	return funcMap, nil
}

func (o *Workflow) GetSchemas() []SchemaDefinition {
	if o.Schemas == nil {
		return make([]SchemaDefinition, 0)
	}

	return o.Schemas
}

func (o *Workflow) GetFunctions() []FunctionDefinition {
	if o.Functions == nil {
		return make([]FunctionDefinition, 0)
	}

	return o.Functions
}

func (o *Workflow) GetFunction(id string) (*FunctionDefinition, error) {

	for i, fn := range o.Functions {
		if fn.ID == id {
			return &o.Functions[i], nil
		}
	}

	return nil, fmt.Errorf("function '%s' not defined", id)

}

func (o *Workflow) UnmarshalYAML(unmarshal func(interface{}) error) error {

	m := make(map[string]interface{})
	err := unmarshal(&m)
	if err != nil {
		return err
	}

	err = o.unmarshal(m)
	if err != nil {
		return err
	}

	return nil

}

func (o *Workflow) Load(data []byte) error {
	return yaml.Unmarshal(data, o)
}

func (o *Workflow) GetStartState() State {

	if o.Start == nil || o.Start.GetState() == "" {
		return o.States[0]
	}

	for _, state := range o.States {
		if state.GetID() == o.Start.GetState() {
			return state
		}
	}

	panic(errors.New("cannot resolve start state"))

}
