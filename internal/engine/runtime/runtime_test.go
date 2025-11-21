package runtime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/google/uuid"
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

	script := `
		function start() {
			return transition(end, "returnValue")
		}	

		function end(payload) {
			log(payload)
			return finish(payload)
		}	
	`

	err := runtime.ExecScript(context.Background(), &runtime.Script{
		InstID:   uuid.New(),
		Text:     script,
		Mappings: "",
		Fn:       "start",
		Input:    "{}",
	}, onFinish, onTransition, runtime.NoOnAction, runtime.NoOnSubflow)
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
			err := runtime.ExecScript(context.Background(), &runtime.Script{
				InstID:   uuid.New(),
				Text:     tt.js,
				Mappings: "",
				Fn:       "start",
				Input:    "{}",
			}, nil, nil, runtime.NoOnAction, runtime.NoOnSubflow)
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
