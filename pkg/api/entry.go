package api

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

func RunApplication(config *core.Config) error {
	s, err := NewServer(config)
	if err != nil {
		return fmt.Errorf("cannot create API server: %w", err)
	}

	err = s.Start
	if err != nil {
		return fmt.Errorf("cannot start API server: %w", err)
	}

	return nil
}
