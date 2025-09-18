package compiler_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/internal/compiler"
)

func TestBody(t *testing.T) {
	tests := []struct {
		script    string
		expectErr bool
	}{
		{
			`
			var flow = { start: "stateOne" };
			function stateOne() {}
			`,
			false,
		},
		{
			`
			var flow1 = { start: "stateOne" };
			function stateOne() {}
			`,
			true,
		},
		{
			`
			var flow = { start: "stateOne" };
			`,
			true,
		},
		{
			`
			function stateOne() {}
			`,
			false,
		},
		{
			`
			function stateThree() {}
			function stateOne() {}
			var flow = { start: "stateOne" };
			function stateOne() {}
			`,
			false,
		},
		{
			`
			var flow = { start: "stateOne" };
			`,
			true,
		},
		{
			`
			function stateOne() {}

			stateOne()
			`,
			true,
		},
		{
			`
			function stateOne() {}

			var g = 1
			`,
			true,
		},
		{
			`
			var flow = { start: "stateOne" };
			var flow = { start: "stateOne" };
			`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			err := compiler.ValidateBody(tt.script, "")
			t.Logf("error %v\n", err)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}

}

func TestReturnValues(t *testing.T) {
	tests := []struct {
		script   string
		expected int
	}{
		{`function anotherNormalFunction() {
			return transition("someState");
		}`, 1},
		{`function normalFunction() {
			return 123;
		}`, 0},
		{`function anotherNormalFunction() {
			return transition("someState");
		}`, 1},
		{`function stateTwo() {
			if (cond) {
				return transition("a");
			}
			return transition("b");
		}`, 0},
		{`function stateOne() {
			if (true) {
				return transition("nextState");
			} else {
				return transition("nextState");
			}
		}`, 0},
		{`function stateOne() {
			if (true) {
				return transition("nextState");
			} else {
				return transition("nextState");
			}

			return false
		  }
		function another() {
			return transition("nextState");
		}`, 2},
		{`function anotherNormalFunction() {
			return finish();
		}`, 1},
		{`function anotherNormalFunction() {
			finish("someState");
		}`, 1},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			errs, _ := compiler.ValidateTransitions(tt.script, "")
			if len(errs) != tt.expected {
				t.Errorf("validate js got errors %d, expected %d (%v)", len(errs), tt.expected, errs)
			}
		})
	}

}

func TestConfig(t *testing.T) {
	tests := []struct {
		script                                     string
		wantType, wantTimeout, wantCron, wantState string
		expectedErr                                bool
	}{
		{`
		function stateOne() { return finish() }
		`, "default", "PT15M", "", "stateOne", false},
		{`
		var flow = {
			cron: "does not work"
		}
		`, "default", "PT15M", "", "", true},
		{`
		var flow = {
			state: "does not exist"
		}
		function stateOne() { return finish() }
		`, "default", "PT15M", "", "stateOne", true},
		{`
		var flow = {
			type: "does not exist"
		}
		`, "default", "PT15M", "", "", true},
		{`
		var flow = {
			timeout: "invalid"
		}
		`, "default", "PT15M", "", "", true},
		{`
		var flow = {
			type: "default",
			timeout: "PT30M"
		}
		function stateOne() { return finish() }
		`, "default", "PT30M", "", "stateOne", false},
		{`
		var flow = {
			type: "default",
			cron: "* * * * *",
			timeout: "PT30M",
			state: "stateOne",
			events: [
				{
					type: "com.github.push",
					context: {
						title: "hello",
						world: "world"
					}
				},
				{
					type: "com.github.pull",
					context: {
						title: "one",
						world: 1223
					}
				}
			]
		}
		function stateOne() { return finish() }
		`, "default", "PT30M", "* * * * *", "stateOne", false},
		{`
		var flow = {
			type: "cron",
		}
		function stateOne() { return finish() }
		`, "cron", "PT15M", "", "stateOne", true},
		{`
		var flow = {
			type: "event",
		}
		function stateOne() { return finish() }
		`, "event", "PT15M", "", "stateOne", true},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			flow, err := compiler.ValidateConfig(tt.script, "")
			if tt.expectedErr {
				if err == nil {
					t.Errorf("validate config got no error, but expected one (%v)", err)
				}

				if err != nil {
					t.Logf("expected failure occurred '%s'", err.Error())
				}
			} else {
				if flow.Type != tt.wantType {
					t.Errorf("type = %q, want %q", flow.Type, tt.wantType)
				}
				if flow.Cron != tt.wantCron {
					t.Errorf("cron = %q, want %q", flow.Cron, tt.wantCron)
				}
				if flow.State != tt.wantState {
					t.Errorf("state = %q, want %q", flow.State, tt.wantState)
				}
				if flow.Timeout != tt.wantTimeout {
					t.Errorf("timeout = %q, want %q", flow.Timeout, tt.wantTimeout)
				}

				b, _ := json.MarshalIndent(flow, "", "   ")
				fmt.Println(string(b))
			}
		})
	}

}
