package model

import (
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

// WorkflowIDRegex - Regex used to validate ID
const WorkflowIDRegex = "^[a-z][a-z0-9._-]{1,34}[a-z0-9]$"

type Workflow struct {
	ID          string               `yaml:"id" json:"id"`
	Name        string               `yaml:"name,omitempty" json:"name,omitempty"`
	Description string               `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string               `yaml:"version,omitempty" json:"version,omitempty"`
	Exclusive   bool                 `yaml:"singular,omitempty" json:"singular,omitempty"`
	Functions   []FunctionDefinition `yaml:"functions,omitempty" json:"functions,omitempty"`
	Schemas     []SchemaDefinition   `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	States      []State              `yaml:"states,omitempty" json:"states,omitempty"`
	Timeouts    *TimeoutDefinition   `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
	Start       StartDefinition      `yaml:"start,omitempty" json:"start,omitempty"`
}

func (o *Workflow) unmarshal(m map[string]interface{}) error {
	// split start out from the rest, and umarshal it
	if err := o.unmStart(m); err != nil {
		return err
	}

	// split states out from the rest
	x, ok := m["states"]
	if !ok {
		return errors.New("states required")
	}

	delete(m, "states")

	if err := strictMapUnmarshal(m, &o); err != nil {
		return fmt.Errorf("failed to decode workflow: %w", err)
	}

	// cast all states
	list, ok := x.([]interface{})
	if !ok {
		return errors.New("invalid type for states")
	}

	o.States = make([]State, len(list))

	for i := range list {
		// insert state in workflow.states[i]
		if err := o.unmState(list[i], i); err != nil {
			return err
		}
	}

	return o.validate()
}

// unmStart - unmarshal "start" object to Workflow
func (o *Workflow) unmStart(m map[string]interface{}) (err error) {
	// split start out from the rest
	y, startFound := m["start"]
	if startFound {
		// Start

		delete(m, "start")
		startMap, startType, err := processInterfaceMap(y)
		if err != nil {
			return fmt.Errorf("bad start: %w", err)
		}

		start, err := getStartFromType(startType)
		if err != nil {
			return fmt.Errorf("start: %w", err)
		}

		if err := strictMapUnmarshal(startMap, &start); err != nil {
			return fmt.Errorf("failed to decode start: %w", err)
		}

		err = start.Validate()
		if err != nil {
			err = fmt.Errorf("start invalid: %w", err)
		}

		o.Start = start
	}

	return err
}

// unmState - unmarshal "state" object to Workflow States
//	the state interface is casted to a supported State 'type'
//	and then inserted into workflow[sIndex]
func (o *Workflow) unmState(state interface{}, sIndex int) error {
	stateMap, stateType, err := processInterfaceMap(state)
	if err != nil {
		return fmt.Errorf("state[%d]: %w", sIndex, err)
	}

	s, err := getStateFromType(stateType)
	if err != nil {
		err = fmt.Errorf("state[%d]: %w", sIndex, err)
	}

	if err := strictMapUnmarshal(stateMap, &s); err != nil {
		return fmt.Errorf("failed to decode state[%d]: %w", sIndex, err)
	}

	o.States[sIndex] = s

	err = s.Validate()
	if err != nil {
		err = fmt.Errorf("state[%d]: %w", sIndex, err)
	}

	return err
}

func (o *Workflow) validate() error {

	if err := o.regexValidateID(); err != nil {
		return err
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
	return o.Timeouts.Validate()
}

func (o *Workflow) regexValidateID() error {
	if o.ID == "" {
		return fmt.Errorf("workflow id required")
	}

	matched, err := regexp.MatchString(WorkflowIDRegex, o.ID)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("workflow ID must match regex: %s", WorkflowIDRegex)
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

// CheckFunctionsInRange - Iterate over defined functions and return error on first function that has a scale larger than param maxScale
func (o *Workflow) CheckFunctionsScaleInRange(maxScale int) error {
	if o.Functions == nil {
		return nil
	}

	for i, f := range o.Functions {
		if f.Scale > maxScale {
			return fmt.Errorf("function[%v] '%s' scale cannot be more than %v", i, f.ID, maxScale)
		}
	}

	return nil
}

func (o *Workflow) GetFunction(id string) (FunctionDefinition, error) {

	for i, fn := range o.Functions {
		if fn.GetID() == id {
			return o.Functions[i], nil
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

// VariableReference - Workflow variable referenced in getter or setter
type VariableReference struct {
	Scope     string   `json:"scope"`
	Key       string   `json:"key"`
	Operation []string `json:"operation"`
}

// GetVariableReferences - Get all varaible referenced in getters & setters from a workflow
func (o *Workflow) GetVariableReferences() []VariableReference {
	refs := make([]VariableReference, 0)
	refsMap := make(map[string]VariableReference)

	// Get All variable references
	for _, state := range o.GetStates() {
		sType := state.GetType()

		// handle variable getter references
		if sType == StateTypeGetter {
			getterState := state.(*GetterState)
			for j := range getterState.Variables {
				key := getterState.Variables[j].Key
				scope := getterState.Variables[j].Scope

				if refEntry, ok := refsMap[key+scope]; ok {
					// Prepend Operation to reference entry
					refEntry.Operation = append([]string{"GET"}, refEntry.Operation...)
					refsMap[key+scope] = refEntry
				} else {
					// Add new reference entry
					refsMap[key+scope] = VariableReference{
						Key:       getterState.Variables[j].Key,
						Scope:     getterState.Variables[j].Scope,
						Operation: []string{"GET"},
					}
				}
			}
		}

		// handle variable setter references
		if sType == StateTypeSetter {
			setterState := state.(*SetterState)
			for j := range setterState.Variables {
				key := setterState.Variables[j].Key
				scope := setterState.Variables[j].Scope

				if refEntry, ok := refsMap[key+scope]; ok {
					// Append Operation to reference entry
					refEntry.Operation = append(refEntry.Operation, "SET")
					refsMap[key+scope] = refEntry
				} else {
					// Add new reference entry
					refsMap[key+scope] = VariableReference{
						Key:       setterState.Variables[j].Key,
						Scope:     setterState.Variables[j].Scope,
						Operation: []string{"SET"},
					}
				}

			}
		}
	}

	// Convert Map to array
	for _, refEntry := range refsMap {
		refs = append(refs, refEntry)
	}

	return refs
}

// GetSecretReferences - Get all secrets referenced in actions
func (o *Workflow) GetSecretReferences() []string {
	refs := make([]string, 0)
	refsMap := make(map[string]bool)

	// Get All secret references
	for _, state := range o.GetStates() {
		sType := state.GetType()

		// handle action secret references
		if sType == StateTypeAction {
			actionState := state.(*ActionState)
			for j := range actionState.Action.Secrets {
				refsMap[actionState.Action.Secrets[j]] = true
			}
		}
	}

	// Convert Map to array
	for secretName := range refsMap {
		refs = append(refs, secretName)
	}

	return refs
}
