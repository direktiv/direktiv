package runtime

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type SecretCommand struct {
	secrets *map[string]string

	rt *Runtime
}

func NewSecretCommand(rt *Runtime, secrets *map[string]string) *SecretCommand {
	return &SecretCommand{
		rt:      rt,
		secrets: secrets,
	}
}

func (sc SecretCommand) GetName() string {
	return "getSecret"
}

func (sc SecretCommand) GetCommandFunction() interface{} {
	return func(in map[string]string) *Secret {
		n, ok := in["name"]
		if !ok {
			throwRuntimeError(sc.rt.vm, DirektivSecretsErrorCode, fmt.Errorf("name for secret not provided"))
		}

		s, ok := (*sc.secrets)[n]
		if !ok {
			throwRuntimeError(sc.rt.vm, DirektivSecretsErrorCode, fmt.Errorf("secret %s does not exist", n))
		}

		return &Secret{
			Value:   s,
			runtime: sc.rt,
		}
	}
}

type Secret struct {
	Value string

	runtime *Runtime
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
