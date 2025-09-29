package engine_test

import (
	"testing"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/stretchr/testify/require"
)

func TestTransition(t *testing.T) {
	vm := sobek.New()

	engine.InjectCommands(vm, uuid.New())

	vm.RunScript("", `
		function start() {
			return transition(end, "returnValue")
		}	

		function end(payload) {
			log(payload)
			return finish(payload)
		}	
	`)

	start, ok := sobek.AssertFunction(vm.Get("start"))
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
		// TODO: Jens, fix this test
		//{
		//	"finish no parameters",
		//	`function start() {
		//		transition(second, "whatever");
		//	}
		//	function second() {
		//		finish();
		//	}
		//	`,
		//},
		//{
		//	"finish too many parameters",
		//	`function start() {
		//		transition(second, "whatever");
		//	}
		//	function second() {
		//		finish(1, 2, 3);
		//	}
		//	`,
		//},
		//{
		//	"finish too many parameters",
		//	`function start() {
		//		return transition(second, "whatever");
		//	}
		//	function second() {
		//		return finish(1, 2, 3);
		//	}
		//	`,
		//},
		//{
		//	"finish error middle stack",
		//	`function start() {
		//		return transition(second, "whatever");
		//	}
		//	function second() {
		//		return transition(second);
		//	}
		//	function third() {
		//		return finish(1, 2, 3);
		//	}
		//	`,
		//},
		//{
		//	"finish error finish stack",
		//	`function start() {
		//		return transition(second, "whatever");
		//	}
		//	function second() {
		//		return transition(third, "");
		//	}
		//	function third() {
		//		return finish(1, 2, 3);
		//	}
		//	`,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := sobek.New()
			engine.InjectCommands(vm, uuid.New())
			vm.RunScript("", tt.js)
			start, ok := sobek.AssertFunction(vm.Get("start"))
			require.True(t, ok)
			_, err := start(sobek.Undefined())
			require.Error(t, err)
		})
	}

}
