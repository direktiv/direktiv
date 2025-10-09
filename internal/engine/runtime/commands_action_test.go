package runtime_test

import (
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/google/uuid"
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
	onFinish := func(output []byte) error {
		gotOutput = output
		return nil
	}
	var gotMemory []string
	onTransition := func(memory []byte, fn string) error {
		gotMemory = append(gotMemory, fmt.Sprintf("%s -> %s", fn, memory))
		return nil
	}

	err = runtime.ExecScript(&runtime.Script{
		InstID:   uuid.New(),
		Text:     script,
		Mappings: "",
		Fn:       ci.Config().Config.State,
		Input:    "{}",
		Metadata: map[string]string{
			"namespace": "test",
		},
	}, onFinish, onTransition)
	require.NoError(t, err)

	require.Equal(t, "\"done\"", string(gotOutput))
}
