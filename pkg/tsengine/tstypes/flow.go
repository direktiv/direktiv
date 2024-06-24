package tstypes

import (
	"fmt"
	"slices"

	"dario.cat/mergo"
)

type TSExecutionContext struct {
	Definition Definition
	Messages   *Messages

	Functions map[string]Function
	ID        string
}

type Function struct {
	Image   string            `json:"image,omitempty"`
	Size    string            `json:"size"`
	Envs    map[string]string `json:"envs,omitempty"`
	Cmd     string            `json:"cmd,omitempty"`
	Init    []string          `json:"init,omitempty"`
	Flow    string            `json:"flow,omitempty"`
	Service string            `json:"service,omitempty"`
}

type Definition struct {
	Type    string  `json:"type,omitempty"`
	State   string  `json:"state,omitempty"`
	Store   string  `json:"store,omitempty"`
	JSON    bool    `json:"json,omitempty"`
	Cron    string  `json:"cron,omitempty"`
	Timeout string  `json:"timeout,omitempty"`
	Scale   []Scale `json:"scale"`
}

func DefaultDefinition() Definition {
	return Definition{
		Type:    defTypeDefault,
		Store:   defStoreAlways,
		JSON:    true,
		Timeout: defTimoutDefault,
		Scale: []Scale{
			{
				Min:    0,
				Max:    1,
				Metric: defMetricInstances,
				Value:  100,
			},
		},
	}
}

type Scale struct {
	Min    int    `json:"min"`
	Max    int    `json:"max,omitempty"`
	Cron   string `json:"cron,omitempty"`
	Metric string `json:"metric,omitempty"`
	Value  int    `json:"value,omitempty"`
}

const (
	defTypeDefault     = "default"
	defTimoutDefault   = "PT15M"
	defStoreAlways     = "always"
	defMetricInstances = "instances"
)

func (def *Definition) Validate() *Messages {
	m := NewMessages()

	return m
}

type Messages struct {
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

func NewMessages() *Messages {
	return &Messages{
		Warnings: make([]string, 0),
		Errors:   make([]string, 0),
	}
}

func (m *Messages) AddError(format string, args ...any) {
	m.Errors = append(m.Errors, fmt.Sprintf(format, args...))
}

func (m *Messages) AddWarning(format string, args ...any) {
	m.Warnings = append(m.Warnings, fmt.Sprintf(format, args...))
}

func (m *Messages) Merge(a *Messages) {
	m.Errors = slices.Concat(m.Errors, a.Errors)
	m.Warnings = slices.Concat(m.Warnings, a.Warnings)
}

func MergeDefinitions(dst, other *Definition) error {
	for i := range dst.Scale {
		err := mergo.Merge(&dst.Scale[i], &other.Scale[0])
		if err != nil {
			return fmt.Errorf("scale %w", err)
		}
	}
	err := mergo.Merge(dst, other)
	if err != nil {
		return fmt.Errorf("definition %w", err)
	}

	return nil
}
