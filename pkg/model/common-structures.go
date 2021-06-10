package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/qri-io/jsonschema"
	"github.com/senseyeio/duration"
)

const CommonNameRegex = "^[a-z][a-z0-9._-]{1,34}[a-z0-9]$"
const VariableNameRegex = `^[^_-][\w]*[^_-]$`

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

type FunctionDefinition struct {
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

	matched, err := regexp.MatchString(CommonNameRegex, o.ID)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("function id must match regex: %s", CommonNameRegex)
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
