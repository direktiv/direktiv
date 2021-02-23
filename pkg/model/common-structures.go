package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/qri-io/jsonschema"
	"github.com/senseyeio/duration"
)

type TimeoutDefinition struct {
	Interrupt string `yaml:"interrupt,omitempty"`
	Kill      string `yaml:"kill,omitempty"`
}

func (o *TimeoutDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.Interrupt != "" && !isISO8601(o.Interrupt) {
		return errors.New("interrupt is not a ISO8601 string")
	}

	if o.Kill != "" && !isISO8601(o.Kill) {
		return errors.New("kill is not a ISO8601 string")
	}
	return nil
}

type FunctionDefinition struct {
	ID    string `yaml:"id"`
	Image string `yaml:"image"`
	Size  Size   `yaml:"size,omitempty"`
	Cmd   string `yaml:"cmd,omitempty"`
}

func (o *FunctionDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.ID == "" {
		return errors.New("id required")
	}

	if o.Image == "" {
		return errors.New("image required")
	}

	return nil

}

type SchemaDefinition struct {
	ID     string      `yaml:"id"`
	Schema interface{} `yaml:"schema"`
}

func (o *SchemaDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.ID == "" {
		return errors.New("id required")
	}

	if err := isJSONSchema(o.Schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	return nil

}

type ActionDefinition struct {
	Function string   `yaml:"function,omitempty"`
	Workflow string   `yaml:"workflow,omitempty"`
	Input    string   `yaml:"input,omitempty"`
	Secrets  []string `yaml:"secrets,omitempty"`
}

func (o *ActionDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.Function != "" && o.Workflow != "" {
		return errors.New("function and workflow cannot coexist")
	}

	if o.Function == "" && o.Workflow == "" {
		return errors.New("must define atleast one function or workflow")
	}

	return nil
}

// utils

func isISO8601(date string) bool {
	_, err := duration.ParseISO8601(date)
	return (err == nil)
}

func isJSONSchema(schema interface{}) error {
	s, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(s, &rs); err != nil {
		return err
	}

	return nil
}

func validateTransformJQ(transform string) error {
	if transform == "" {
		return nil
	}

	if _, err := gojq.Parse(transform); err != nil {
		return fmt.Errorf("transform is an invalid jq string: %v", err)
	}

	return nil
}
