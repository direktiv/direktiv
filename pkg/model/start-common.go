package model

import "errors"

type StartDefinition interface {
	GetState() string
	GetType() StartType
	Validate() error
	GetEvents() []StartEventDefinition
}

func (o *Workflow) GetStartDefinition() StartDefinition {

	if o.Start != nil {
		return o.Start
	}

	return &DefaultStart{}

}

// FIXME: Going to be renamed later
type StartEventDefinition struct {
	Type    string                 `yaml:"type"`
	Filters map[string]interface{} `yaml:"filters,omitempty"`
}

func (o *StartEventDefinition) Validate() error {
	if o.Type == "" {
		return errors.New("type required")
	}

	return nil
}

type StartCommon struct {
	Type  StartType `yaml:"type"`
	State string    `yaml:"state,omitempty"`
}

func (o *StartCommon) commonValidate() error {
	// if o.Type == "" {
	// 	return errors.New("type required")
	// }
	return nil
}

func (o *StartCommon) GetType() StartType {

	if o == nil {
		return StartTypeDefault
	}

	return o.Type

}

func (o *StartCommon) GetState() string {

	if o == nil {
		return ""
	}

	return o.State

}
