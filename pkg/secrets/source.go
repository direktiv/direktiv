package secrets

import (
	"context"
	"errors"
	"fmt"
)

const (
	DefaultSourceString = "default"
)

var ErrSecretNotFound = NewJSONMarshalableError(errors.New("secret not found"))

type SourceConfig struct {
	Name   string
	Driver string
	Data   []byte
}

type Source interface {
	Get(ctx context.Context, path string) ([]byte, error)
}

type NullSource struct {
	Name string
}

func (s *NullSource) Get(_ context.Context, path string) ([]byte, error) {
	return nil, NewJSONMarshalableError(fmt.Errorf("source not found: %s", s.Name))
}

type NullDriverSource struct {
	Name string
}

func (s *NullDriverSource) Get(_ context.Context, path string) ([]byte, error) {
	return nil, NewJSONMarshalableError(fmt.Errorf("driver not found: %s", s.Name))
}

type BadConfigSource struct {
	Name  string
	Error error
}

func (s *BadConfigSource) Get(_ context.Context, path string) ([]byte, error) {
	return nil, NewJSONMarshalableError(fmt.Errorf("invalid config for '%s' driver: %w", s.Name, s.Error))
}
