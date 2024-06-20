package tsservice

import "encoding/base64"

type Secret struct {
	Name string
}

func (s *Secret) Validate() *Messages {
	m := newMessages()
	if s.Name == "" {
		m.addError("secret requires a name")
	}
	return m
}

func (s *Secret) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(s.Name))
}
