package tstypes

import (
	"fmt"
	"slices"
)

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

func (m *Messages) Merge(a *Messages) {
	m.Errors = slices.Concat(m.Errors, a.Errors)
	m.Warnings = slices.Concat(m.Warnings, a.Warnings)
}
