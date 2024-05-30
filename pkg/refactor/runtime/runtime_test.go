package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
	"github.com/direktiv/direktiv/pkg/refactor/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createRuntime(t *testing.T, script string, json bool) *runtime.Runtime {

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

	s := make(map[string]string)
	fn := make(map[string]string)

	rt, err := runtime.New(id, c.Program, &s, &fn, f, json)
	assert.NoError(t, err)

	return rt
}
