package runtime

import (
	"testing"

	// "github.com/direktiv/direktiv/internal/engine"

	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/stretchr/testify/require"
)

func TestTransition(t *testing.T) {
	rt := New(uuid.New(), map[string]string{}, "")

	_, err := rt.RunScript("", `
		function start() {
			return transition(end, "returnValue")
		}	

		function end(payload) {
			log(payload)
			return finish(payload)
		}	
	`)
	require.NoError(t, err)

	start, ok := sobek.AssertFunction(rt.vm.Get("start"))
	require.True(t, ok)

	g, err := start(sobek.Undefined())
	require.NoError(t, err)

	require.Equal(t, "returnValue", g.Export())
}

func TestTransitionErrors(t *testing.T) {
	tests := []struct {
		name string
		js   string
	}{
		{
			"transition no parameters",
			`function start() {
				transition();
			}`,
		},
		{
			"transition one parameters",
			`function start() {
				transition("noFunction");
			}`,
		},
		{
			"transition two parameters wrong",
			`function start() {
				transition("noFunction", "whatever");
			}`,
		},
		{
			"transition two parameters not exist",
			`function start() {
				transition(doesNotExist, "whatever");
			}`,
		},
		{
			"transition two parameters wrong type",
			`function start() {
				transition(second, "whatever");
			}
			var second = "";
			`,
		},
		{
			"transition two parameters wrong type",
			`function start() {
				transition(second, "whatever");
			}
			function second() {
				ssdsd;
			}
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := New(uuid.New(), map[string]string{}, "")
			rt.RunScript("", tt.js)
			start, ok := sobek.AssertFunction(rt.vm.Get("start"))
			require.True(t, ok)
			_, err := start(sobek.Undefined())
			require.Error(t, err)
		})
	}

}
