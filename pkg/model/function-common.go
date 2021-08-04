package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// -------- Function Types -----------

type FunctionType int

const (
	DefaultFunctionType           FunctionType = iota
	ReusableContainerFunctionType              // Old school knative
	IsolatedContainerFunctionType              // isolated (scale field not needed)
	NamespacedKnativeFunctionType
	GlobalKnativeFunctionType
	SubflowFunctionType
	// Namespace Level Knative services (no size, command or scale), image should be name of service
	// Global Level Knative services
	// Subflow Level // id, workflow
)

var FunctionTypeStrings = []string{"unknown", "reusable", "isolated", "knative-namespace", "knative-global", "subflow"}

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

unknown:

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

type FunctionDefinition interface {
	GetID() string
	GetType() FunctionType
	Validate() error
}

type FunctionFileDefinition struct {
	Key   string `yaml:"key" json:"key"`
	As    string `yaml:"as,omitempty" json:"as,omitempty"`
	Scope string `yaml:"scope,omitempty" json:"scope,omitempty"`
	Type  string `yaml:"type,omitempty" json:"type,omitempty"`
}

func (o FunctionFileDefinition) Validate() error {

	if o.Key == "" {
		return errors.New("key required")
	}

	switch o.Scope {
	case "":
	case "namespace":
	case "workflow":
	case "instance":
	default:
		return errors.New("bad scope (choose 'namespace', 'workflow', or 'instance')")
	}

	switch o.Type {
	case "":
	case "plain":
	case "base64":
	case "tar":
	case "tar.gz":
	default:
		return errors.New("bad type (choose 'plain', 'base64', 'tar', or 'tar.gz'")
	}

	return nil

}

// util
func getFunctionDefFromType(ftype string) (FunctionDefinition, error) {
	var f FunctionDefinition
	var err error

	switch ftype {
	case ReusableContainerFunctionType.String():
		f = new(ReusableFunctionDefinition)
	case IsolatedContainerFunctionType.String():
		f = new(IsolatedFunctionDefinition)
	case NamespacedKnativeFunctionType.String():
		f = new(NamespacedFunctionDefinition)
	case GlobalKnativeFunctionType.String():
		f = new(GlobalFunctionDefinition)
	case SubflowFunctionType.String():
		f = new(SublowFunctionDefinition)
	case "":
		err = errors.New("type required")
	default:
		err = errors.New("type unimplemented/unrecognized")
	}

	return f, err
}
