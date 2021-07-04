package model

import (
	"errors"
)

type ErrorState struct {
	StateCommon `yaml:",inline"`
	Error       string      `yaml:"error"`
	Message     string      `yaml:"message"`
	Args        []string    `yaml:"args,omitempty"`
	Transform   interface{} `yaml:"transform,omitempty"`
	Transition  string      `yaml:"transition,omitempty"`
}

func (o *ErrorState) GetID() string {
	return o.ID
}

func (o *ErrorState) getTransitions() map[string]string {
	transitions := make(map[string]string)
	if o.Transition != "" {
		transitions["transition"] = o.Transition
	}

	return transitions
}

func (o *ErrorState) GetTransitions() []string {
	transitions := make([]string, 0)
	if o.Transition != "" {
		transitions = append(transitions, o.Transition)
	}

	return transitions
}

func (o *ErrorState) GetArgs() []string {
	if o.Args == nil {
		return make([]string, 0)
	}

	return o.Args
}

func (o *ErrorState) Validate() error {
	if err := o.commonValidate(); err != nil {
		return err
	}

	if s, ok := o.Transform.(string); ok {
		if err := validateTransformJQ(s); err != nil {
			return err
		}
	}

	if o.Error == "" {
		return errors.New("error required")
	}

	if o.Message == "" {
		return errors.New("message required")
	}

	return nil
}
