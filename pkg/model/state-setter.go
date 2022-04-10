package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/direktiv/direktiv/pkg/util"
)

const DefaultVarMimeType = "application/json"
const RegexVarMimeType = `\w+\/[-+.\w]+`

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
	MimeType string      `yaml:"mimeType,omitempty"`
}

func (a *SetterDefinition) UnmarshalJSON(data []byte) error {

	type SetterDefinitionAlias SetterDefinition

	var s SetterDefinitionAlias

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	// Set Default
	if s.MimeType == "" {
		s.MimeType = DefaultVarMimeType
	}

	// Set Definition
	a.Key = s.Key
	a.Scope = s.Scope
	a.Value = s.Value
	a.MimeType = s.MimeType

	return nil

}

func (a *SetterDefinition) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var s interface{}

	err := unmarshal(&s)
	if err != nil {
		return err
	}

	sD := s.(SetterDefinition)

	// Set Default
	if sD.MimeType == "" {
		sD.MimeType = DefaultVarMimeType
	}

	*a = sD

	return nil
}

func (o *SetterDefinition) Validate() error {

	match, err := regexp.MatchString(RegexVarMimeType, o.MimeType)
	if err != nil {
		return errors.New(`regex validation of mime type failed`)
	}

	if !match {
		return errors.New(`mimeType is not a valid MIME type string`)
	}

	switch o.Scope {
	case "instance":
	case "workflow":
	case "namespace":
	case "thread":
	default:
		return ErrVarScope
	}

	if o.Key == "" {
		return errors.New(`key required`)
	}

	if !util.VarNameRegex.MatchString(o.Key) {
		return fmt.Errorf("key is invalid: must start with a letter and only contain letters, numbers and '_'")
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
