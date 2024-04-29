package states

import (
	"fmt"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/jqer"
)

const (
	ErrCodeJQBadQuery                 = "direktiv.jq.badCommand"
	ErrCodeJQNoResults                = "direktiv.jq.badCommand"
	ErrCodeJQManyResults              = "direktiv.jq.badCommand"
	ErrCodeJQNotObject                = "direktiv.jq.notObject"
	ErrCodeFailedSchemaValidation     = "direktiv.schema.failed"
	ErrCodeJQNotString                = "direktiv.jq.notString"
	ErrCodeInvalidVariableKey         = "direktiv.var.invalidKey"
	ErrCodeInvalidVariableScope       = "direktiv.var.invalidScope"
	ErrCodeAllBranchesFailed          = "direktiv.parallel.allFailed"
	ErrCodeNotArray                   = "direktiv.foreach.badArray"
	ErrCodeInvalidVariablePermissions = "direktiv.var.perms"
)

func wrap(err error, s string) error {
	return fmt.Errorf(s, err)
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

	if len(output) == 0 {
		return nil, derrors.NewCatchableError(ErrCodeJQNoResults, "the `jq` or `js` command produced no outputs")
	}

	if len(output) != 1 {
		return nil, derrors.NewCatchableError(ErrCodeJQManyResults, "the `jq` or `js` command produced multiple outputs")
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

func jqString(input interface{}, command interface{}) (string, error) {
	x, err := jqOne(input, command)
	if err != nil {
		return "", err
	}

	s, ok := x.(string)
	if !ok {
		s = fmt.Sprintf("%v", x)
	}

	return s, nil
}

func truth(x interface{}) bool {
	var success bool

	if x != nil { //nolint:nestif
		switch v := x.(type) {
		case bool:
			if v {
				success = true
			}
		case string:
			if v != "" {
				success = true
			}
		case int:
			if v != 0 {
				success = true
			}
		case float64:
			if v != 0.0 {
				success = true
			}
		case []interface{}:
			if len(v) > 0 {
				success = true
			}
		case map[string]interface{}:
			if len(v) > 0 {
				success = true
			}
		default:
		}
	}

	return success
}
