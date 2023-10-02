package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

type ActionDefinition struct {
	Function string                   `yaml:"function,omitempty"`
	Input    interface{}              `yaml:"input,omitempty"`
	Secrets  []string                 `yaml:"secrets,omitempty"`
	Retries  *RetryDefinition         `yaml:"retries,omitempty"`
	Files    []FunctionFileDefinition `json:"files,omitempty"    yaml:"files,omitempty"`
}

func (o *ActionDefinition) Validate() error {
	if o == nil {
		return nil
	}

	if o.Function == "" {
		return errors.New("must define at least one function or workflow")
	}

	if o.Retries != nil {
		err := o.Retries.Validate()
		if err != nil {
			return err
		}
	}

	for i, f := range o.Files {
		err := f.Validate()
		if err != nil {
			return fmt.Errorf("function file %d: %w", i, err)
		}
	}

	return nil
}

// utils

func isISO8601(date string) bool {
	_, err := duration.ParseISO8601(date)
	return err == nil
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

func processInterfaceMap(s interface{}) (map[string]interface{}, string, error) {
	var iType string

	iMap, ok := s.(map[string]interface{})
	if !ok {
		return iMap, iType, errors.New("invalid")
	}

	iT, ok := iMap["type"]
	if !ok {
		return iMap, iType, fmt.Errorf("missing 'type' field")
	}

	iType, ok = iT.(string)
	if !ok {
		return iMap, iType, fmt.Errorf("bad data-format for 'type' field")
	}

	return iMap, iType, nil
}

func strictMapUnmarshal(m map[string]interface{}, target interface{}) error {
	// unmarshal top level fields into Workflow
	data, err := json.Marshal(&m)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err) // This error should be impossible
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force Unknown fields to throw error

	if err := dec.Decode(&target); err != nil {
		return errors.New(strings.TrimPrefix(err.Error(), "json: "))
	}

	return nil
}
