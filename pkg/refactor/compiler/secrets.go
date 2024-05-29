package compiler

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
