package mirror

import (
	"os"
)

// MockedSource mocks Source interface.
type MockedSource struct {
	Paths map[string]string
}

var _ Source = &MockedSource{}

func (m MockedSource) PullInPath(mirrorSettings Settings, distDirectory string) error {
	for k, v := range m.Paths {
		//nolint
		if err := os.WriteFile(distDirectory+k, []byte(v), 0o644); err != nil {
			return err
		}
	}

	return nil
}
