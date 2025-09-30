package compiler_test

import (
	"sync"
	"testing"

	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/stretchr/testify/assert"
)

func TestTranspilerLoop(t *testing.T) {

	script := `
		function stateOne(payload) {
			print("RUN STATE FIRST");
    		return transition(stateTwo, payload);
		}
		function stateTwo(payload) {
			print("RUN STATE SECOND");
    		return finish(payload);
		}`

	var wg sync.WaitGroup
	for range make([]int, 20) {
		wg.Go(func() {
			ci := compiler.NewCompileItem([]byte(script), "")
			err := ci.TranspileAndValidate()
			assert.NoError(t, err)
		})
	}

	wg.Wait()
}

func TestTranspiler(t *testing.T) {

	tt, err := compiler.NewTranspiler()
	assert.NoError(t, err)

	_, _, err = tt.Transpile("const hallo = \"world\"", "dummy")
	assert.NoError(t, err)

	script := `const flow : FlowDefintion = {
		json: false
	}

	jens()
	function start(state) {
		const f = new FlowFile({
			name: "input.data"
		})

		return f.base64()
	}

	function jens() {
		let a = 10
	}
	`

	_, _, err = tt.Transpile(script, "dummy")

	assert.NoError(t, err)
}
