package runtime_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTransition(t *testing.T) {
	var gotOutput []byte
	var onFinish runtime.OnFinishHook = func(output []byte) error {
		gotOutput = output
		return nil
	}
	var gotMemory []string
	var onTransition runtime.OnTransitionHook = func(memory []byte, fn string) error {
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
	}, onFinish, onTransition)
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
			})
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
			t.Fatalf("ParseFuncName(%q) = %q; want %q", tc.in, tc.want, tc.want)
		}
	}
}

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple string", "hello world"},
		{"empty string", ""},
		{"numbers", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := fmt.Sprintf("%q", tt.input)
			encoded := base64.StdEncoding.EncodeToString([]byte(tt.input))
			expected = fmt.Sprintf("%q", encoded)

			script := fmt.Sprintf(`
				function start() {
					return finish(base64Encode("%s"));
				}
			`, tt.input)

			var gotOutput []byte
			var onFinish runtime.OnFinishHook = func(output []byte) error {
				gotOutput = output
				return nil
			}

			err := runtime.ExecScript(context.Background(), &runtime.Script{
				InstID:   uuid.New(),
				Text:     script,
				Mappings: "",
				Fn:       "start",
				Input:    "{}",
			}, onFinish)
			require.NoError(t, err)
			require.Equal(t, expected, string(gotOutput))
		})
	}
}

func TestBase64Decode(t *testing.T) {
	tests := []struct {
		name    string
		decoded string
	}{
		{"simple string", "hello world"},
		{"empty string", ""},
		{"numbers", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := base64.StdEncoding.EncodeToString([]byte(tt.decoded))
			expected := fmt.Sprintf("%q", tt.decoded)

			script := fmt.Sprintf(`
				function start() {
					return finish(base64Decode("%s"));
				}
			`, encoded)

			var gotOutput []byte
			var onFinish runtime.OnFinishHook = func(output []byte) error {
				gotOutput = output
				return nil
			}

			err := runtime.ExecScript(context.Background(), &runtime.Script{
				InstID:   uuid.New(),
				Text:     script,
				Mappings: "",
				Fn:       "start",
				Input:    "{}",
			}, onFinish)
			require.NoError(t, err)
			require.Equal(t, expected, string(gotOutput))
		})
	}
}

func TestBase64DecodeInvalid(t *testing.T) {
	script := `
		function start() {
			return finish(base64Decode("not-valid!!!"));
		}
	`

	err := runtime.ExecScript(context.Background(), &runtime.Script{
		InstID:   uuid.New(),
		Text:     script,
		Mappings: "",
		Fn:       "start",
		Input:    "{}",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid base64")
}

func TestBase64Roundtrip(t *testing.T) {
	testInputs := []string{
		"hello world",
		"12345",
	}

	for _, input := range testInputs {
		t.Run(input, func(t *testing.T) {
			script := fmt.Sprintf(`
				function start() {
					let encoded = base64Encode("%s");
					let decoded = base64Decode(encoded);
					return finish(decoded);
				}
			`, input)

			var gotOutput []byte
			var onFinish runtime.OnFinishHook = func(output []byte) error {
				gotOutput = output
				return nil
			}

			err := runtime.ExecScript(context.Background(), &runtime.Script{
				InstID:   uuid.New(),
				Text:     script,
				Mappings: "",
				Fn:       "start",
				Input:    "{}",
			}, onFinish)
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("%q", input), string(gotOutput))
		})
	}
}
