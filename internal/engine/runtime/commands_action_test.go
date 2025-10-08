package runtime_test

import (
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/stretchr/testify/require"
)

func TestActionParsing(t *testing.T) {
	script := `
	var myAction = generateAction({
	type: "local",
	size: "medium",
	image: "my/image",
	envs: {
			my: "value",
			hello: "world",
			eins: "zwei",
			"200": "kjjj"
		}
	});

	function stateOne(payload) {
		myAction();
		return finish("done");
	}
	`

	ci := compiler.NewCompileItem([]byte(script), "")
	err := ci.TranspileAndValidate()

	require.NoError(t, err)
	fmt.Println(ci.ValidationErrors)

	var gotOutput []byte
	recordOutput := func(output []byte) error {
		gotOutput = output
		return nil
	}

	rt := runtime.New(uuid.New(), map[string]string{}, "", recordOutput)
	_, err = rt.RunScript("", script)
	require.NoError(t, err)

	start, ok := sobek.AssertFunction(rt.GetVar("stateOne"))
	require.True(t, ok)

	_, err = start(sobek.Undefined())
	require.NoError(t, err)
	require.Equal(t, "\"done\"", string(gotOutput))
}
