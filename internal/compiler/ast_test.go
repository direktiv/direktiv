package compiler_test

import (
	"testing"

	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/stretchr/testify/require"
)

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
			parser, err := compiler.NewASTParser(tt.script, "")
			require.NoError(t, err)

			parser.ValidateTransitions()
			if len(parser.Errors) != tt.expected {
				t.Errorf("validate js got errors %d, expected %d (%v)", len(parser.Errors), tt.expected, parser.Errors)
			}
		})
	}

}

func TestConfig(t *testing.T) {
	transpiler, _ := compiler.NewTranspiler()

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
			script, mapping, err := transpiler.Transpile(tt.script, "dummy")
			require.NoError(t, err)

			parser, err := compiler.NewASTParser(script, mapping)
			require.NoError(t, err)

			flow, err := parser.ValidateConfig()
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
			}
		})
	}

}

func TestVariables(t *testing.T) {
	transpiler, _ := compiler.NewTranspiler()

	tests := []struct {
		name     string
		ts       string
		errCount int
	}{
		{
			"allowed secrets call but not with function call",
			`
			secrets("my-secret-key");
			secrets(getValues());
			`,
			1,
		},
		{
			"functions direct not allowed",
			`
			console.log("this should fail");
			alert("this should fail");
			setTimeout(function() {}, 1000);
			Math.random();
			`,
			4,
		},
		{
			"normal vars allowed",
			`
			var x = 5;
			var y = "hello";
			var z = x + y;
			var arr = [1, 2, 3];
			var obj = {a: 1, b: "test"};
			var bool = true;
			var nullVar = null;
			`,
			0,
		},
		{
			"function calls in variable assignments not allowed",
			`
			var badVar1 = console.log("fail");
			var badVar2 = Math.random();
			var badVar3 = setTimeout(function() {}, 1000);
			var badVar5 = JSON.parse("{}");
			var badVar6 = parseInt("123");
			var badVar7 = parseFloat("123.45");
			`,
			6,
		},
		{
			"new is allowed but no function",
			`
			let good1 = new MyThing();
			let good2 = new MyThing({
			a: 1, b: 100});
			let bad1 = new MyThing({
			a: 1, b: getValue()});
			const bad2 = new MyThing(doSomething());
			`,
			2,
		},
		{
			"method calls on objects not allowed",
			`
			var badVar8 = someObject.method();
			var badVar9 = arr.push(4);
			var badVar10 = str.toLowerCase();
			var badVar11 = obj.toString();
			`,
			4,
		},
		{
			"chained method calls not allowed",
			`
			var badVar12 = someObject.method().anotherMethod();
			var badVar13 = arr.filter(x => x > 0).map(x => x * 2);
			`,
			2,
		},
		{
			"function calls with property access not allowed",
			`
			var badVar14 = window.alert("fail");
			var badVar15 = document.getElementById("test");
			var badVar16 = global.require("module");
			`,
			3,
		},
		{
			"function calls in complex expressions not allowed",
			`
			var badVar21 = getValue() + 5;
			var badVar22 = 10 + Math.random();
			var badVar23 = someFunction() || "default";
			var badVar24 = condition ? getValue() : "default";
			var badVar25 = !isEmpty();
			`,
			5,
		},
		{
			"function calls in array literals not allowed",
			`
			var badVar26 = [getValue(), 1, 2];
			var badVar27 = [Math.random(), Math.random()]; // two errors
			`,
			3,
		},
		{
			"function calls in object literals not allowed",
			`
			var badVar28 = {a: getValue(), b: 2};
			var badVar29 = {timestamp: Date.now(), value: 1};
			var badVar30 = {id: generateId(), name: "test"};
			`,
			3,
		},
		{
			"nested function calls not allowed",
			`
			var badVar31 = outerFunction(innerFunction());
			var badVar32 = Math.max(getValue(), getOtherValue());
			`,
			2,
		},
		{
			"function calls with computed property access not allowed",
			`
			var badVar33 = obj[getKey()];
			var badVar34 = arr[getIndex()];
			`,
			2,
		},
		{
			"callback functions that call other functions not allowed",
			`
			var badVar35 = [1,2,3].map(function(x) { return transform(x); });
			`,
			1,
		},
		{
			"immediately invoked function expression not allowed",
			`
			var badVar36 = (function() { return getValue(); })();
			`,
			1,
		},
		{
			"spread operator with function calls not allowed",
			`
			var badVar38 = [...getArray()];
			var badVar39 = {...getObject()};
			`,
			2,
		},
		{
			"destructuring with function calls not allowed",
			`
			var {a, b} = getObject();
			var [first, second] = getArray();
			`,
			2,
		},
		{
			"function calls in binary expressions not allowed",
			`
			var badVar40 = getValue() === expected;
			var badVar41 = getCount() > 0;
			var badVar42 = getName() + " suffix";
			`,
			3,
		},
		{
			"function calls in unary expressions not allowed",
			`
			var badVar43 = +getString();
			var badVar44 = typeof getValue();
			var badVar45 = delete obj[getKey()];
			`,
			3,
		},
		{
			"function calls with this context not allowed",
			`
			var badVar46 = this.method();
			var badVar47 = self.getValue();
			`,
			2,
		},
		{
			"generator function calls not allowed",
			`
			var badVar49 = getGenerator().next();
			`,
			1,
		},
		{
			"function calls in switch statements not allowed",
			`
			var badVar50 = getValue() ? 1 : 2;
			`,
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, mapping, err := transpiler.Transpile(tt.ts, "dummy")
			require.NoError(t, err)

			parser, err := compiler.NewASTParser(script, mapping)
			require.NoError(t, err)
			parser.ValidateFunctionCalls()

			if len(parser.Errors) != tt.errCount {
				t.Errorf("script '%s' = errors %d, want %d", tt.name, len(parser.Errors), tt.errCount)
			}

			for i := range parser.Errors {
				ee := parser.Errors[i]
				t.Logf("error '%s' (line: %d, column: %d)", ee.Message, ee.Line, ee.Column)
			}
		})
	}
}
