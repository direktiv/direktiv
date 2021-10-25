package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vorteil/direktiv/pkg/util"
)

type VarMimeType int

const (
	VarMimeTypeJSON VarMimeType = iota
	VarMimeTypePlainText
	VarMimeTypeOctetStream
)

var VarMimeTypeStrings = []string{"application/json", "text/plain", "application/octet-stream"}

type SetterState struct {
	StateCommon `yaml:",inline"`
	Variables   []SetterDefinition `yaml:"variables"`
	Transform   interface{}        `yaml:"transform,omitempty"`
	Transition  string             `yaml:"transition,omitempty"`
}

type SetterDefinition struct {
	Scope    string      `yaml:"scope,omitempty"`
	Key      string      `yaml:"key"`
	Value    interface{} `yaml:"value,omitempty"`
	MimeType VarMimeType `yaml:"mimeType,omitempty"`
}

func (a VarMimeType) String() string {
	return VarMimeTypeStrings[a]
}

func ParseVarMimeType(s string) (VarMimeType, error) {

	if s == "" {
		goto unknown
	}

	for i := range VarMimeTypeStrings {
		if VarMimeTypeStrings[i] == s {
			return VarMimeType(i), nil
		}
	}

unknown:

	return VarMimeType(0), fmt.Errorf("unrecognized mime type (should be one of [%s]): %s", strings.Join(VarMimeTypeStrings, ", "), s)

}

func (a VarMimeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *VarMimeType) UnmarshalJSON(data []byte) error {

	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	x, err := ParseVarMimeType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

func (a VarMimeType) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

func (a *VarMimeType) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var s string

	err := unmarshal(&s)
	if err != nil {
		return err
	}

	x, err := ParseVarMimeType(s)
	if err != nil {
		return err
	}

	*a = x

	return nil

}

func (o *SetterDefinition) Validate() error {

	if o.Scope == "" {
		return errors.New(`scope required ("instance", "workflow", or "namespace")`)
	}

	switch o.Scope {
	case "instance":
	case "workflow":
	case "namespace":
	default:
		return fmt.Errorf(`invalid scope '%s' (requires "instance", "workflow", or "namespace")`, o.Scope)
	}

	if o.Key == "" {
		return errors.New(`key required`)
	}

	if ok := util.MatchesVarRegex(o.Key); !ok {
		return fmt.Errorf("variable key must match regex: %s", util.RegexPattern)
	}

	if o.Value == "" {
		return errors.New(`value required`)
	}

	if s, ok := o.Value.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	return nil

}

func (o *SetterState) GetID() string {
	return o.ID
}

func (o *SetterState) getTransitions() map[string]string {
	transitions := make(map[string]string)
	if o.Transition != "" {
		transitions["transition"] = o.Transition
	}

	for i, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions[fmt.Sprintf("errors[%v]", i)] = errDef.Transition
		}
	}

	return transitions
}

func (o *SetterState) GetTransitions() []string {
	transitions := make([]string, 0)
	if o.Transition != "" {
		transitions = append(transitions, o.Transition)
	}

	for _, errDef := range o.ErrorDefinitions() {
		if errDef.Transition != "" {
			transitions = append(transitions, errDef.Transition)
		}
	}

	return transitions
}

func (o *SetterState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if len(o.Variables) == 0 {
		return errors.New("variables required")
	}

	for i, varDef := range o.Variables {
		if err := varDef.Validate(); err != nil {
			return fmt.Errorf("variables[%d] is invalid: %v", i, err)
		}
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %v", i, err)
		}
	}

	return nil
}
