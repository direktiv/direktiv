package runtime

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Secret struct {
	Value string

	runtime *Runtime
}

func (rt *Runtime) getSecret(in map[string]string) *Secret {

	n, ok := in["name"]
	if !ok {
		throwRuntimeError(rt.vm, DirektivSecretsErrorCode, fmt.Errorf("name for secret not provided"))
	}

	s, ok := rt.manager.RuntimeData().Secrets[n]
	if !ok {
		throwRuntimeError(rt.vm, DirektivSecretsErrorCode, fmt.Errorf("secret %s does not exist", n))
	}

	return &Secret{
		Value:   s,
		runtime: rt,
	}
}

func (s *Secret) String() string {
	return s.Value
}

func (s *Secret) File(name string, perm int) {
	p := filepath.Join(s.runtime.dirInfo().instanceDir, name)
	err := os.WriteFile(p, []byte(s.Value), fs.FileMode(perm))
	if err != nil {
		throwRuntimeError(s.runtime.vm, DirektivFileErrorCode, err)
	}
}

func (s *Secret) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(s.Value))
}
