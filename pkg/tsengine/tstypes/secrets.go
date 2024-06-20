package tstypes

import (
	"encoding/base64"
)

type Secret struct {
	Name string
}

func (s *Secret) Validate() *Messages {
	m := NewMessages()
	if s.Name == "" {
		m.AddError("secret requires a name")
	}

	return m
}

func (s *Secret) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(s.Name))
}
