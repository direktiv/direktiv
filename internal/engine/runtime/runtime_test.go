package runtime_test

import (
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/stretchr/testify/require"
)

func TestTransition(t *testing.T) {
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

	start, ok := sobek.AssertFunction(rt.GetVar("start"))
	require.True(t, ok)
	_, err = start(sobek.Undefined())
	require.NoError(t, err)

	require.Equal(t, "\"returnValue\"", string(gotOutput))
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
			rt := runtime.New(uuid.New(), map[string]string{}, "", runtime.NoOnFinish, runtime.NoOnTransition)
			rt.RunScript("", tt.js)
			start, ok := sobek.AssertFunction(rt.GetVar("start"))
			require.True(t, ok)
			_, err := start(sobek.Undefined())
			require.Error(t, err)
		})
	}
}

func TestParseFuncName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"function stateTwo(payload)", "stateTwo"},
		{"function   myFunc()", "myFunc"},
		{"  function  spaced  (x, y)", "spaced"},
		{"function _private(arg)", "_private"},
		{"function name_with_digits123(a)", "name_with_digits123"},
		{"function unicodeŁódź(x)", "unicodeŁódź"}, // allowed by our simple splitter
		{"notAFunction something()", ""},
		{"function noParen", ""},
		{"", ""},
		{"function (x)", ""}, // empty name before '('
	}

	for _, tc := range tests {
		got := runtime.ParseFuncNameFromText(tc.in)
		if got != tc.want {
			t.Fatalf("ParseFuncName(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}
