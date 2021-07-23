package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// -------------- Branch Modes --------------

type BranchMode int

const (
	BranchModeAnd BranchMode = iota
	BranchModeOr
)

var branchModeStrings []string = []string{
	"and",
	"or",
}

func ParseBranchMode(s string) (BranchMode, error) {

	if s == "" {
		return 0, fmt.Errorf("mode must be one of %v", branchModeStrings)
	}

	for i, str := range branchModeStrings {
		if str == s {
			return BranchMode(i), nil
		}
	}

	return 0, fmt.Errorf("unknown mode '%s' (must be one of %v)", s, branchModeStrings)

}

func (a BranchMode) String() string {
	return branchModeStrings[a]
}

func (a BranchMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *BranchMode) UnmarshalJSON(data []byte) error {

	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	x, err := ParseBranchMode(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

func (a BranchMode) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

func (a *BranchMode) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var s string

	err := unmarshal(&s)
	if err != nil {
		return err
	}

	x, err := ParseBranchMode(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

// -------------- Size --------------

// Size string enum to differentiate function sizes
type Size int

const (
	SmallSize Size = iota
	MediumSize
	LargeSize
)

var sizeStrings []string = []string{
	"small",
	"medium",
	"large",
}

func ParseSize(s string) (Size, error) {

	if s == "" {
		return 0, fmt.Errorf("size must be one of %v", sizeStrings)
	}

	for i, str := range sizeStrings {
		if str == s {
			return Size(i), nil
		}
	}

	return 0, fmt.Errorf("unknown size '%s' (must be one of %v)", s, sizeStrings)

}

func (a Size) String() string {
	return sizeStrings[a]
}

func (a Size) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Size) UnmarshalJSON(data []byte) error {

	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	x, err := ParseSize(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

func (a Size) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

func (a *Size) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var s string

	err := unmarshal(&s)
	if err != nil {
		return err
	}

	x, err := ParseSize(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

// -------------- State Types --------------

type StateType int

const (
	StateTypeAction StateType = iota
	StateTypeConsumeEvent
	StateTypeDelay
	StateTypeEventsAnd
	StateTypeEventsXor
	StateTypeError
	StateTypeForEach
	StateTypeGenerateEvent
	StateTypeNoop
	StateTypeParallel
	StateTypeSwitch
	StateTypeValidate
	StateTypeConsume
	StateTypeCallback
	StateTypeGetter
	StateTypeSetter
)

var stateTypeStrings []string = []string{
	"action",
	"consumeEvent",
	"delay",
	"eventAnd",
	"eventXor",
	"error",
	"foreach",
	"generateEvent",
	"noop",
	"parallel",
	"switch",
	"validate",
	"consumeEvent",
	"callback",
	"getter",
	"setter",
}

func ParseStateType(s string) (StateType, error) {

	if s == "" {
		return 0, fmt.Errorf("type must be one of %v", stateTypeStrings)
	}

	for i, str := range stateTypeStrings {
		if str == s {
			return StateType(i), nil
		}
	}

	return 0, fmt.Errorf("unknown type '%s' (must be one of %v)", s, stateTypeStrings)

}

func (a StateType) String() string {
	return stateTypeStrings[a]
}

func (a StateType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *StateType) UnmarshalJSON(data []byte) error {

	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	x, err := ParseStateType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

func (a StateType) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

func (a *StateType) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var s string

	err := unmarshal(&s)
	if err != nil {
		return err
	}

	x, err := ParseStateType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

// -------------- Start Types --------------

type StartType int

const (
	StartTypeDefault StartType = iota
	StartTypeScheduled
	StartTypeEvent
	StartTypeEventsXor
	StartTypeEventsAnd
)

var startTypeStrings []string = []string{
	"default",
	"scheduled",
	"event",
	"eventsXor",
	"eventsAnd",
}

func ParseStartType(s string) (StartType, error) {

	if s == "" {
		return 0, fmt.Errorf("type must be one of %v", startTypeStrings)
	}

	for i, str := range startTypeStrings {
		if str == s {
			return StartType(i), nil
		}
	}

	return 0, fmt.Errorf("unknown type '%s' (must be one of %v)", s, startTypeStrings)

}

func (a StartType) String() string {
	return startTypeStrings[a]
}

func (a StartType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *StartType) UnmarshalJSON(data []byte) error {

	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	x, err := ParseStartType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

func (a StartType) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

func (a *StartType) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var s string

	err := unmarshal(&s)
	if err != nil {
		return err
	}

	x, err := ParseStartType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

// -------- Function Types -----------

type FunctionType int

const (
	DefaultFunctionType FunctionType = iota
	ReusableContainerFunctionType
	IsolatedContainerFunctionType
)

var FunctionTypeStrings = []string{"default", "reusable", "isolated"}

func (a FunctionType) String() string {
	return FunctionTypeStrings[a]
}

func ParseFunctionType(s string) (FunctionType, error) {

	if s == "" {
		return FunctionType(0), nil
	}

	for i := range FunctionTypeStrings {
		if FunctionTypeStrings[i] == s {
			return FunctionType(i), nil
		}
	}

	return FunctionType(0), fmt.Errorf("unrecognized function type (should be one of [%s]): %s", strings.Join(FunctionTypeStrings, ", "), s)

}

func (a FunctionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *FunctionType) UnmarshalJSON(data []byte) error {

	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	x, err := ParseFunctionType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

func (a FunctionType) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

func (a *FunctionType) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var s string

	err := unmarshal(&s)
	if err != nil {
		return err
	}

	x, err := ParseFunctionType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}
