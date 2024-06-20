package deleteme

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/direktiv/direktiv/pkg/tsengine/commands"
	"github.com/direktiv/direktiv/pkg/tsengine/runtime"
	"github.com/direktiv/direktiv/pkg/tsengine/state"
	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createRuntime(t *testing.T, s, fn map[string]string, script string, json bool) *state.Executor {

	c, err := tsservice.New("dummy", script)
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

	err = rb.WithCommand(&commands.BtoaCommand{})
	assert.NoError(t, err)
	err = rb.WithCommand(&commands.SleepCommand{})
	assert.NoError(t, err)
	err = rb.WithCommand(&commands.AtobCommand{})
	assert.NoError(t, err)
	err = rb.WithCommand(&commands.TrimCommand{})
	assert.NoError(t, err)
	err = rb.WithCommand(&commands.ToJSONCommand{})
	assert.NoError(t, err)
	err = rb.WithCommand(&commands.FromJSONCommand{})
	assert.NoError(t, err)
	lc, err := commands.NewLogCommand([]interface{}{
		"test", "testv"})
	assert.NoError(t, err)
	err = rb.WithCommand(lc)
	assert.NoError(t, err)
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
