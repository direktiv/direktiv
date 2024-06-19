package transpiler_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/transpiler"
	"github.com/stretchr/testify/assert"
)

func TestTranspiler(t *testing.T) {

	tt, _ := transpiler.NewTranspiler()

	_, err := tt.Transpile("const hallo = \"world\"")
	assert.NoError(t, err)

	script := `
	const flow : FlowDefintion = {
		json: false
	}
	
	function start(state) {
		const f = new FlowFile({
			name: "input.data"
		})
		return f.base64()
	}`

	_, err = tt.Transpile(script)
	assert.NoError(t, err)
}
