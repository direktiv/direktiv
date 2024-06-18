package deleteme

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/direktiv/direktiv/pkg/commands"
	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/direktiv/direktiv/pkg/state"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createRuntime(t *testing.T, s, fn map[string]string, script string, json bool) *state.Executor {

	c, err := compiler.New("dummy", script)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	f, err := os.MkdirTemp("", "test")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	id := uuid.New()

	// usually done by caller
	os.MkdirAll(filepath.Join(f, "instances", id.String()), 0777)
	os.MkdirAll(filepath.Join(f, "shared", id.String()), 0777)

	rb, err := runtime.New(id, f, json)

	err = rb.WithCommand(commands.NewFileCommand(rb))
	assert.NoError(t, err)
	err = rb.WithCommand(commands.NewSecretCommand(rb, &s))
	assert.NoError(t, err)
	err = rb.WithCommand(commands.NewFunctionCommand(rb, &fn))
	assert.NoError(t, err)
	err = rb.WithCommand(commands.NewRequestCommand(rb))
	assert.NoError(t, err)

	rt, err := state.New(rb, c.Program)
	assert.NoError(t, err)

	return rt
}
