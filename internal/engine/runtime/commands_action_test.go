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
		image: "python",
		envs: {
				my: "value",
				hello: "world",
				eins: "zwei",
				"200": "kjjj"
			}
		});

		function stateOne(payload) {
			myAction({ 
				data: "mydata",
				files: "dsds"
			});
			return finish("done");
		}
	`

	ci := compiler.NewCompileItem([]byte(script), "")
	err := ci.TranspileAndValidate()

	require.NoError(t, err)
	fmt.Println(ci.ValidationErrors)

	var gotOutput []byte
	onFinish := func(output []byte) error {
		gotOutput = output
		return nil
	}
	var gotMemory []string
	onTransition := func(memory []byte, fn string) error {
		gotMemory = append(gotMemory, fmt.Sprintf("%s -> %s", fn, memory))
		return nil
	}

	rt := runtime.New(uuid.New(), map[string]string{}, "", onFinish, onTransition)
	_, err = rt.RunScript("", script)
	require.NoError(t, err)

	start, ok := sobek.AssertFunction(rt.GetVar("stateOne"))
	require.True(t, ok)

	_, err = start(sobek.Undefined())
	require.NoError(t, err)
	require.Equal(t, "\"done\"", string(gotOutput))
}
