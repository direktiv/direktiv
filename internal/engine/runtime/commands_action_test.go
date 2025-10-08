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

	rt := runtime.New(uuid.New(), map[string]string{}, "")
	_, err = rt.RunScript("", script)
	require.NoError(t, err)

	start, ok := sobek.AssertFunction(rt.GetVar("stateOne"))
	require.True(t, ok)

	v, err := start(sobek.Undefined())
	require.NoError(t, err)

	fmt.Println(v)

}
