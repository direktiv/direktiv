package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Secret struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Source string `json:"source"`
	Data   []byte `json:"data"`
	Error  error  `json:"error"`
}

type List []Secret

func (l List) Len() int {
	return len(l)
}

func (l List) Less(i, j int) bool {
	a := l[i]
	b := l[j]

	// sort by source first
	if a.Source < b.Source {
		return true
	}

	if b.Source < a.Source {
		return false
	}

	// sort by path second
	return a.Path < b.Path
}

func (l List) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type jsonMarshalableError struct {
	err error
}

func (e *jsonMarshalableError) Error() string {
	return e.err.Error()
}

func (e *jsonMarshalableError) Unwrap() error {
	return errors.Unwrap(e.err)
}

func (e *jsonMarshalableError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

func NewJSONMarshalableError(err error) error {
	return &jsonMarshalableError{
		err: err,
	}
}

type SecretRef struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Source string `json:"source"`
}

func (r *SecretRef) UnmarshalJSON(data []byte) error {
	var v interface{}

	r.Name, r.Path, r.Source = "", "", ""

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch x := v.(type) {
	case map[string]interface{}:
		if s, err := extractString(&x, "name"); err == nil {
			r.Name = s
		} else {
			return err
		}

		if s, err := extractString(&x, "path"); err == nil {
			r.Path = s
		} else {
			return err
		}

		if s, err := extractString(&x, "source"); err == nil {
			r.Source = s
		} else {
			return err
		}

		for k := range x {
			return fmt.Errorf("unexpected sub-field in secret reference: '%s'", k)
		}
	case string:
		r.Name = x
	default:
		return errors.New("invalid json type for secret reference")
	}

	return nil
}

func extractString(m *map[string]interface{}, key string) (string, error) {
	var r string

	if a, defined := (*m)[key]; defined {
		if s, ok := a.(string); ok {
			r = s
		} else {
			return "", fmt.Errorf("invalid json type for secret reference sub-field '%s'", key)
		}

		delete((*m), key)
	}

	return r, nil
}

func (r *SecretRef) Validate() error {
	if r.Name == "" {
		return errors.New("secret reference name is empty")
	}

	return nil
}
