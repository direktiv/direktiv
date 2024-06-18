package commands

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/runtime"
)

type SecretCommand struct {
	secrets *map[string]string

	rt *runtime.Runtime
}

func NewSecretCommand(rt *runtime.Runtime, secrets *map[string]string) *SecretCommand {
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
			runtime.ThrowRuntimeError(sc.rt.VM, runtime.DirektivSecretsErrorCode, fmt.Errorf("name for secret not provided"))
		}

		s, ok := (*sc.secrets)[n]
		if !ok {
			runtime.ThrowRuntimeError(sc.rt.VM, runtime.DirektivSecretsErrorCode, fmt.Errorf("secret %s does not exist", n))
		}

		return &Secret{
			Value:   s,
			runtime: sc.rt,
		}
	}
}

type Secret struct {
	Value string

	runtime *runtime.Runtime
}

func (s *Secret) String() string {
	return s.Value
}

func (s *Secret) File(name string, perm int) {
	p := filepath.Join(s.runtime.DirInfo().InstanceDir, name)
	err := os.WriteFile(p, []byte(s.Value), fs.FileMode(perm))
	if err != nil {
		runtime.ThrowRuntimeError(s.runtime.VM, runtime.DirektivFileErrorCode, err)
	}
}

func (s *Secret) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(s.Value))
}
