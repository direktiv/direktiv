package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// -------- Function Types -----------

type FunctionType int

const (
	DefaultFunctionType FunctionType = iota
	ReusableContainerFunctionType
	IsolatedContainerFunctionType
)

var FunctionTypeStrings = []string{"unknown", "reusable", "isolated"}

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
	GetType() StateType
	Validate() error
}

type ReusableFunctionDefinition struct {
	Type  FunctionType             `yaml:"type"`
	ID    string                   `yaml:"id"`
	Image string                   `yaml:"image"`
	Size  Size                     `yaml:"size,omitempty"`
	Cmd   string                   `yaml:"cmd,omitempty"`
	Scale int                      `yaml:"scale,omitempty"`
	Files []FunctionFileDefinition `yaml:"files,omitempty"`
}

func (o *FunctionDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.ID == "" {
		return errors.New("id required")
	}

	matched, err := regexp.MatchString(FunctionNameRegex, o.ID)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("function id must match regex: %s", FunctionNameRegex)
	}

	if o.Image == "" {
		return errors.New("image required")
	}

	for i, f := range o.Files {
		err := f.Validate()
		if err != nil {
			return fmt.Errorf("function file %d: %v", i, err)
		}
	}

	return nil

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
