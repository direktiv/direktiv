package mirror

import (
	"os"
)

// MockedSource mocks Source interface.
type MockedSource struct {
	Paths map[string]string
}

var _ Source = &MockedSource{} // Ensures MockedSource struct conforms to Source interface.

func (m MockedSource) PullInPath(settings Settings, dst string) error {
	for k, v := range m.Paths {
		if err := os.WriteFile(dst+k, []byte(v), 0o600); err != nil {
			return err
		}
	}

	return nil
}
