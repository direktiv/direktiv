package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// -------- Function Types -----------
//
//nolint:recvcheck
type FunctionType int

const (
	DefaultFunctionType           FunctionType = iota
	ReusableContainerFunctionType              // Old school knative
	NamespacedKnativeFunctionType
	SystemKnativeFunctionType
	SubflowFunctionType
)

var FunctionTypeStrings = []string{"unknown", "knative-workflow", "knative-namespace", "knative-system", "subflow"}

func (a FunctionType) String() string {
	return FunctionTypeStrings[a]
}

func ParseFunctionType(s string) (FunctionType, error) {
	if s == "" {
		goto unknown
	}

	for i := range FunctionTypeStrings {
		if FunctionTypeStrings[i] == s {
			return FunctionType(i), nil
		}
	}

	if s == "reusable" {
		return ReusableContainerFunctionType, nil
	}

unknown:

	return FunctionType(0), fmt.Errorf("unrecognized function type (should be one of [%s]): %s", strings.Join(FunctionTypeStrings[1:], ", "), s)
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

type FunctionDefinition interface {
	GetID() string
	GetType() FunctionType
	Validate() error
}

type FunctionFileDefinition struct {
	Key         string `json:"key"                   yaml:"key"`
	As          string `json:"as,omitempty"          yaml:"as,omitempty"`
	Scope       string `json:"scope,omitempty"       yaml:"scope,omitempty"`
	Type        string `json:"type,omitempty"        yaml:"type,omitempty"`
	Permissions string `json:"permissions,omitempty" yaml:"permissions,omitempty"`
}

var (
	ErrVarScope    = errors.New(`bad scope (choose 'namespace', 'workflow', 'thread', 'instance', 'system', or 'file')`)
	ErrVarReadOnly = errors.New(`'file' scope variables cannot be written to from workflows'`)
	ErrVarNotFile  = errors.New("target is not a file")
)

func (o FunctionFileDefinition) Validate() error {
	if o.Key == "" {
		return errors.New("key required")
	}

	return nil
}

func getFunctionDefFromType(ftype string) (FunctionDefinition, error) {
	var f FunctionDefinition
	var err error

	switch ftype {
	case "reusable":
		fallthrough
	case ReusableContainerFunctionType.String():
		f = new(ReusableFunctionDefinition)
	case NamespacedKnativeFunctionType.String():
		f = new(NamespacedFunctionDefinition)
	case SystemKnativeFunctionType.String():
		f = new(SystemFunctionDefinition)
	case SubflowFunctionType.String():
		f = new(SubflowFunctionDefinition)
	case "":
		err = errors.New("type required(reusable, knative-workflow, knative-namespace, knative-system, subflow)")
	default:
		err = errors.New("type unrecognized(reusable, knative-workflow, knative-namespace, knative-system, subflow)")
	}

	return f, err
}
