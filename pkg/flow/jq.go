package flow

import (
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/jqer"
)

func (srv *server) initJQ() {
	jqer.StringQueryRequiresWrappings = true
	jqer.TrimWhitespaceOnQueryStrings = true

	jqer.SearchInStrings = true
	jqer.WrappingBegin = "jq"
	jqer.WrappingIncrement = "("
	jqer.WrappingDecrement = ")"
}

func jq(input interface{}, command interface{}) ([]interface{}, error) {
	out, err := jqer.Evaluate(input, command)
	if err != nil {
		return nil, derrors.NewCatchableError(ErrCodeJQBadQuery, "failed to evaluate jq/js: %v", err)
	}

	return out, nil
}

func jqOne(input interface{}, command interface{}) (interface{}, error) {
	output, err := jq(input, command)
	if err != nil {
		return nil, err
	}

	if len(output) != 1 {
		return nil, derrors.NewCatchableError(ErrCodeJQNotObject, "the `jq` or `js` command produced multiple outputs")
	}

	return output[0], nil
}

func jqObject(input interface{}, command interface{}) (map[string]interface{}, error) {
	x, err := jqOne(input, command)
	if err != nil {
		return nil, err
	}

	m, ok := x.(map[string]interface{})
	if !ok {
		return nil, derrors.NewCatchableError(ErrCodeJQNotObject, "the `jq` or `js` command produced a non-object output")
	}

	return m, nil
}
