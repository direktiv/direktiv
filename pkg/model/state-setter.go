package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/utils"
)

const (
	DefaultVarMimeType = "application/json"
	RegexVarMimeType   = `\w+\/[-+.\w]+`
)

type SetterState struct {
	StateCommon `yaml:",inline"`
	Variables   []SetterDefinition `yaml:"variables"`
	Transform   interface{}        `yaml:"transform,omitempty"`
	Transition  string             `yaml:"transition,omitempty"`
}

type SetterDefinition struct {
	Scope    string      `yaml:"scope,omitempty"`
	Key      interface{} `yaml:"key"`
	Value    interface{} `yaml:"value,omitempty"`
	MimeType interface{} `yaml:"mimeType,omitempty"`
}

func (o *SetterDefinition) UnmarshalJSON(data []byte) error {
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
	o.Key = s.Key
	o.Scope = s.Scope
	o.Value = s.Value
	o.MimeType = s.MimeType

	return nil
}

func (o *SetterDefinition) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

	*o = sD

	return nil
}

func (o *SetterDefinition) Validate() error {
	switch o.Scope {
	case utils.VarScopeInstance:
	case utils.VarScopeWorkflow:
	case utils.VarScopeNamespace:
	case utils.VarScopeThread:
	case utils.VarScopeFileSystem:
		return ErrVarReadOnly
	default:
		return ErrVarScope
	}

	if o.Key == nil || o.Key == "" {
		return errors.New(`key required`)
	}

	if o.Value == "" {
		return errors.New(`value required`)
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
			return fmt.Errorf("variables[%d] is invalid: %w", i, err)
		}
	}

	for i, errDef := range o.ErrorDefinitions() {
		if err := errDef.Validate(); err != nil {
			return fmt.Errorf("catch[%v] is invalid: %w", i, err)
		}
	}

	return nil
}
