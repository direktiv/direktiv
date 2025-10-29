package compiler_test

import (
	"testing"

	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/stretchr/testify/require"
)

// TestStateFunctionValidation tests that state functions must return transition/finish
func TestStateFunctionValidation(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		errCount int
	}{
		{
			name: "state function with valid transition return",
			script: `function stateOne() {
				return transition("nextState");
			}`,
			errCount: 0,
		},
		{
			name: "state function with valid finish return",
			script: `function stateOne() {
				return finish();
			}`,
			errCount: 0,
		},
		{
			name: "state function with conditional returns",
			script: `function stateTwo() {
				if (cond) {
					return transition("a");
				}
				return transition("b");
			}`,
			errCount: 0,
		},
		{
			name: "state function without return",
			script: `function stateOne() {
				let x = 5;
			}`,
			errCount: 1,
		},
		{
			name: "state function with invalid return",
			script: `function stateOne() {
				return "not a transition";
			}`,
			errCount: 1,
		},
		{
			name: "state function with mixed returns",
			script: `function stateOne() {
				if (true) {
					return transition("nextState");
				}
				return false;
			}`,
			errCount: 1,
		},
		{
			name: "multiple state functions, one invalid",
			script: `
			function stateOne() {
				return transition("two");
			}
			function stateTwo() {
				return "invalid";
			}`,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := compiler.NewASTParser(tt.script, "")
			require.NoError(t, err)

			err = parser.Parse()
			require.NoError(t, err)

			if len(parser.Errors) != tt.errCount {
				t.Errorf("got %d errors, want %d", len(parser.Errors), tt.errCount)
				for _, e := range parser.Errors {
					t.Logf("  error: %s (line %d:%d)", e.Message, e.StartLine, e.StartColumn)
				}
			}
		})
	}
}

// TestNonStateFunctionValidation tests that non-state functions cannot use transition/finish
func TestNonStateFunctionValidation(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		errCount int
	}{
		{
			name: "regular function with normal return",
			script: `function normalFunction() {
				return 123;
			}`,
			errCount: 0,
		},
		{
			name: "regular function returning transition",
			script: `function normalFunction() {
				return transition("someState");
			}`,
			errCount: 1,
		},
		{
			name: "regular function returning finish",
			script: `function normalFunction() {
				return finish();
			}`,
			errCount: 1,
		},
		{
			name: "regular function calling transition",
			script: `function normalFunction() {
				transition("someState");
			}`,
			errCount: 1,
		},
		{
			name: "regular function calling finish",
			script: `function normalFunction() {
				finish();
			}`,
			errCount: 1,
		},
		{
			name: "regular function with transition in if statement",
			script: `function helper() {
				if (condition) {
					return transition("state");
				}
				return true;
			}`,
			errCount: 1,
		},
		{
			name: "mixed state and non-state functions",
			script: `
			function stateOne() {
				return transition("two");
			}
			function helper() {
				return transition("invalid");
			}`,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := compiler.NewASTParser(tt.script, "")
			require.NoError(t, err)

			err = parser.Parse()
			require.NoError(t, err)

			if len(parser.Errors) != tt.errCount {
				t.Errorf("got %d errors, want %d", len(parser.Errors), tt.errCount)
				for _, e := range parser.Errors {
					t.Logf("  error: %s (line %d:%d)", e.Message, e.StartLine, e.StartColumn)
				}
			}
		})
	}
}

// TestFirstStateFunction tests that the first state function is correctly identified
func TestFirstStateFunction(t *testing.T) {
	tests := []struct {
		name          string
		script        string
		expectedState string
	}{
		{
			name: "single state function",
			script: `function stateOne() {
				return finish();
			}`,
			expectedState: "stateOne",
		},
		{
			name: "multiple state functions, first is selected",
			script: `
			function stateOne() { return finish(); }
			function stateTwo() { return finish(); }
			function stateThree() { return finish(); }`,
			expectedState: "stateOne",
		},
		{
			name: "state function after regular function",
			script: `
			function helper() { return true; }
			function stateMain() { return finish(); }`,
			expectedState: "stateMain",
		},
		{
			name: "no state function",
			script: `
			function helper() { return true; }
			function another() { return false; }`,
			expectedState: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := compiler.NewASTParser(tt.script, "")
			require.NoError(t, err)

			err = parser.Parse()
			require.NoError(t, err)

			if parser.FirstStateFunc != tt.expectedState {
				t.Errorf("got first state function %q, want %q", parser.FirstStateFunc, tt.expectedState)
			}
		})
	}
}

// TestFlowConfig tests flow configuration parsing and validation
func TestFlowConfig(t *testing.T) {
	transpiler, _ := compiler.NewTranspiler()

	tests := []struct {
		name        string
		script      string
		wantType    string
		wantTimeout string
		wantCron    string
		wantState   string
		expectError bool
	}{
		{
			name: "default flow without config",
			script: `
			function stateOne() { return finish(); }`,
			wantType:    "default",
			wantTimeout: "PT15M",
			wantCron:    "",
			wantState:   "stateOne",
			expectError: false,
		},
		{
			name: "invalid cron pattern",
			script: `
			var flow = {
				cron: "does not work"
			}
			function stateOne() { return finish(); }`,
			wantType:    "default",
			wantTimeout: "PT15M",
			expectError: true,
		},
		{
			name: "state reference to non-existent function",
			script: `
			var flow = {
				state: "doesNotExist"
			}
			function stateOne() { return finish(); }`,
			expectError: true,
		},
		{
			name: "invalid flow type",
			script: `
			var flow = {
				type: "invalid"
			}
			function stateOne() { return finish(); }`,
			expectError: true,
		},
		{
			name: "invalid timeout pattern",
			script: `
			var flow = {
				timeout: "invalid"
			}
			function stateOne() { return finish(); }`,
			expectError: true,
		},
		{
			name: "custom timeout",
			script: `
			var flow = {
				type: "default",
				timeout: "PT30M"
			}
			function stateOne() { return finish(); }`,
			wantType:    "default",
			wantTimeout: "PT30M",
			wantState:   "stateOne",
			expectError: false,
		},
		{
			name: "complete flow config with events",
			script: `
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
					}
				]
			}
			function stateOne() { return finish(); }`,
			wantType:    "default",
			wantTimeout: "PT30M",
			wantCron:    "* * * * *",
			wantState:   "stateOne",
			expectError: false,
		},
		{
			name: "cron type without cron pattern",
			script: `
			var flow = {
				type: "cron"
			}
			function stateOne() { return finish(); }`,
			expectError: true,
		},
		{
			name: "event type without events",
			script: `
			var flow = {
				type: "event"
			}
			function stateOne() { return finish(); }`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, mapping, err := transpiler.Transpile(tt.script, "dummy")
			require.NoError(t, err)

			parser, err := compiler.NewASTParser(script, mapping)
			require.NoError(t, err)

			err = parser.Parse()

			if tt.expectError {
				if len(parser.Errors) == 0 && err == nil {
					t.Errorf("expected error but got none")
				} else {
					t.Logf("got expected error: %v, errors: %d", err, len(parser.Errors))
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, parser.FlowConfig)

			if parser.FlowConfig.Type != tt.wantType {
				t.Errorf("type = %q, want %q", parser.FlowConfig.Type, tt.wantType)
			}
			if parser.FlowConfig.Timeout != tt.wantTimeout {
				t.Errorf("timeout = %q, want %q", parser.FlowConfig.Timeout, tt.wantTimeout)
			}
			if parser.FlowConfig.Cron != tt.wantCron {
				t.Errorf("cron = %q, want %q", parser.FlowConfig.Cron, tt.wantCron)
			}
			if parser.FlowConfig.State != tt.wantState {
				t.Errorf("state = %q, want %q", parser.FlowConfig.State, tt.wantState)
			}
		})
	}
}

// TestTopLevelFunctionCalls tests that only secrets/getSecrets/generateAction are allowed at top level
func TestTopLevelFunctionCalls(t *testing.T) {
	transpiler, _ := compiler.NewTranspiler()

	tests := []struct {
		name     string
		script   string
		errCount int
	}{
		{
			name: "secrets call allowed",
			script: `
			const sec = getSecrets("my-secret-key");
			function stateOne() { return finish(); }`,
			errCount: 0,
		},
		{
			name: "generateAction call allowed",
			script: `
			generateAction({
				type: "local",
				image: "ubuntu",
				cmd: "echo hello"
			});
			function stateOne() { return finish(); }`,
			errCount: 0,
		},
		{
			name: "invalid function call at top level",
			script: `
			console.log("this should fail");
			function stateOne() { return finish(); }`,
			errCount: 1,
		},
		{
			name: "multiple invalid function calls",
			script: `
			console.log("fail");
			alert("fail");
			setTimeout(function() {}, 1000);
			Math.random();
			function stateOne() { return finish(); }`,
			errCount: 4,
		},
		{
			name: "function call in variable assignment",
			script: `
			var badVar = console.log("fail");
			function stateOne() { return finish(); }`,
			errCount: 1,
		},
		{
			name: "nested function calls",
			script: `
			var badVar = outerFunction(innerFunction());
			function stateOne() { return finish(); }`,
			errCount: 2, // Both outer and inner function calls are not allowed
		},
		{
			name: "function calls in arrays",
			script: `
			var arr = [getValue(), 1, 2];
			function stateOne() { return finish(); }`,
			errCount: 1,
		},
		{
			name: "function calls in objects",
			script: `
			var obj = {a: getValue(), b: 2};
			function stateOne() { return finish(); }`,
			errCount: 1,
		},
		{
			name: "method calls not allowed",
			script: `
			var badVar = someObject.method();
			function stateOne() { return finish(); }`,
			errCount: 1,
		},
		{
			name: "function calls inside functions are allowed",
			script: `
			function stateOne() {
				console.log("this is allowed inside function");
				const x = Math.random();
				return finish();
			}`,
			errCount: 0,
		},
		{
			name: "secrets with function call argument",
			script: `
			getSecrets(getValues());
			function stateOne() { return finish(); }`,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, mapping, err := transpiler.Transpile(tt.script, "dummy")
			require.NoError(t, err)

			parser, err := compiler.NewASTParser(script, mapping)
			require.NoError(t, err)

			err = parser.Parse()
			require.NoError(t, err)

			if len(parser.Errors) != tt.errCount {
				t.Errorf("got %d errors, want %d", len(parser.Errors), tt.errCount)
				for _, e := range parser.Errors {
					t.Logf("  error: %s (line %d:%d)", e.Message, e.StartLine, e.StartColumn)
				}
			}
		})
	}
}

// TestGenerateActionCollection tests that all generateAction calls are found and parsed
func TestGenerateActionCollection(t *testing.T) {
	transpiler, _ := compiler.NewTranspiler()

	tests := []struct {
		name         string
		script       string
		expectCount  int
		expectImages []string
	}{
		{
			name: "single generateAction at top level",
			script: `
			generateAction({
				type: "local",
				image: "ubuntu:latest",
				cmd: "echo hello"
			});
			function stateOne() { return finish(); }`,
			expectCount:  1,
			expectImages: []string{"ubuntu:latest"},
		},
		{
			name: "multiple generateAction calls",
			script: `
			generateAction({
				type: "local",
				image: "alpine",
				cmd: "ls"
			});
			generateAction({
				type: "local",
				image: "debian",
				cmd: "pwd"
			});
			function stateOne() { return finish(); }`,
			expectCount:  2,
			expectImages: []string{"alpine", "debian"},
		},
		{
			name: "generateAction inside function",
			script: `
			function stateOne() {
				generateAction({
					type: "local",
					image: "nginx",
					cmd: "start"
				});
				return finish();
			}`,
			expectCount:  1,
			expectImages: []string{"nginx"},
		},
		{
			name: "no generateAction",
			script: `
			function stateOne() { return finish(); }`,
			expectCount:  0,
			expectImages: []string{},
		},
		{
			name: "generateAction with non-local type ignored",
			script: `
			generateAction({
				type: "namespace",
				image: "ubuntu"
			});
			function stateOne() { return finish(); }`,
			expectCount: 0,
		},
		{
			name: "generateAction with environment variables",
			script: `
			generateAction({
				type: "local",
				image: "ubuntu",
				cmd: "env",
				envs: [
					{name: "VAR1", value: "value1"},
					{name: "VAR2", value: "value2"}
				]
			});
			function stateOne() { return finish(); }`,
			expectCount:  1,
			expectImages: []string{"ubuntu"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, mapping, err := transpiler.Transpile(tt.script, "dummy")
			require.NoError(t, err)

			parser, err := compiler.NewASTParser(script, mapping)
			require.NoError(t, err)

			err = parser.Parse()
			require.NoError(t, err)

			if len(parser.Actions) != tt.expectCount {
				t.Errorf("got %d actions, want %d", len(parser.Actions), tt.expectCount)
			}

			for i, action := range parser.Actions {
				if i < len(tt.expectImages) {
					if action.Image != tt.expectImages[i] {
						t.Errorf("action[%d].Image = %q, want %q", i, action.Image, tt.expectImages[i])
					}
				}
			}
		})
	}
}

// TestComplexScenarios tests complete workflows with multiple features
func TestComplexScenarios(t *testing.T) {
	transpiler, _ := compiler.NewTranspiler()

	tests := []struct {
		name          string
		script        string
		expectErrors  int
		expectActions int
		expectState   string
	}{
		{
			name: "complete valid workflow",
			script: `
			const secrets = getSecrets("db-creds");
			
			generateAction({
				type: "local",
				image: "postgres",
				cmd: "psql"
			});
			
			var flow = {
				type: "default",
				timeout: "PT1H",
				state: "stateInit"
			};
			
			function stateInit() {
				console.log("Starting");
				return transition("stateProcess");
			}
			
			function stateProcess() {
				const result = doWork();
				return finish();
			}
			
			function helper() {
				return "helper result";
			}`,
			expectErrors:  0,
			expectActions: 1,
			expectState:   "stateInit",
		},
		{
			name: "workflow with validation errors",
			script: `
			// Invalid function call at top level
			console.log("fail");
			
			var flow = {
				type: "cron"
				// Missing cron pattern - will error
			};
			
			// State function without return
			function stateOne() {
				let x = 5;
			}
			
			// Regular function using transition
			function helper() {
				return transition("invalid");
			}`,
			expectErrors:  4, // console.log + missing cron + no return + helper using transition
			expectActions: 0,
			expectState:   "stateOne",
		},
		{
			name: "nested function scenarios",
			script: `
			function stateOne() {
				const fn = function() {
					// Nested function can call anything
					return someFunction();
				};
				
				const arrow = () => {
					// Arrow function also can call anything
					doSomething();
				};
				
				return finish();
			}`,
			expectErrors:  0,
			expectActions: 0,
			expectState:   "stateOne",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, mapping, err := transpiler.Transpile(tt.script, "dummy")
			require.NoError(t, err)

			parser, err := compiler.NewASTParser(script, mapping)
			require.NoError(t, err)

			_ = parser.Parse()

			if len(parser.Errors) != tt.expectErrors {
				t.Errorf("got %d errors, want %d", len(parser.Errors), tt.expectErrors)
				for _, e := range parser.Errors {
					t.Logf("  error: %s (line %d:%d)", e.Message, e.StartLine, e.StartColumn)
				}
			}

			if len(parser.Actions) != tt.expectActions {
				t.Errorf("got %d actions, want %d", len(parser.Actions), tt.expectActions)
			}

			if parser.FirstStateFunc != tt.expectState {
				t.Errorf("got first state %q, want %q", parser.FirstStateFunc, tt.expectState)
			}
		})
	}
}
