package secrets_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/direktiv/direktiv/pkg/secrets"
)

type MockSource struct {
	Secrets map[string][]byte
}

func (s *MockSource) Get(_ context.Context, path string) ([]byte, error) {
	data, defined := s.Secrets[path]
	if !defined {
		return nil, secrets.ErrSecretNotFound
	}

	return data, nil
}

type MockSourceDriver struct {
}

func (d *MockSourceDriver) ConstructSource(data []byte) secrets.Source {
	s := new(MockSource)

	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}

	return s
}

func TestMockSourceDriver(t *testing.T) {
	_ = secrets.RegisterDriver("mock", &MockSourceDriver{})

	x := MockSource{
		Secrets: map[string][]byte{
			"x": []byte("5"),
		},
	}

	data, _ := json.Marshal(x)

	d, err := secrets.GetDriver("mock")
	if err != nil {
		t.Error(err)
	}

	s := d.ConstructSource(data)
	a, err := s.Get(context.Background(), "x")
	if err != nil {
		t.Error(err)
	}

	if "5" != string(a) {
		t.Errorf("unexpected result")
	}
}
