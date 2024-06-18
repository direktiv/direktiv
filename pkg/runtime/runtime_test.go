package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createRuntime(t *testing.T, s, fn map[string]string, script string, json bool) *runtime.Runtime {

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

	err = rb.WithCommand(runtime.NewFileCommand(rb))
	assert.NoError(t, err)
	err = rb.WithCommand(runtime.NewSecretCommand(rb, &s))
	assert.NoError(t, err)
	err = rb.WithCommand(runtime.NewFunctionCommand(rb, &fn))
	assert.NoError(t, err)
	err = rb.WithCommand(runtime.NewRequestCommand(rb))
	assert.NoError(t, err)

	rt, err := rb.Prepare(c.Program)
	assert.NoError(t, err)

	return rt
}
