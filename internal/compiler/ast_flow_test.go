package compiler_test

import (
	"testing"

	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/stretchr/testify/require"
)

var flow = `
var flow = {
	state: "stateOne"
}

function stateUnused(data) {
	return finish("done")
}

function stateOne() {
	if (true) {
		return transition(stateSecond, "")
	} else {
		return finish("done")
	}
}

function stateSecond(data) {
	return transition(stateThird, "")
}

function stateThird(data) {
	return finish("done")
}
`

func TestFlowchart(t *testing.T) {

	ci := compiler.NewCompileItem([]byte(flow), "/test.wf.ts")
	err := ci.TranspileAndValidate()
	require.NoError(t, err)

	require.Len(t, ci.Config().Config.StateViews, 4)
	require.True(t, ci.Config().Config.StateViews["stateUnused"].Finish)
	require.Len(t, ci.Config().Config.StateViews["stateOne"].Transitions, 1)
}
