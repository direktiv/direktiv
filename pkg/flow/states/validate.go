package states

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/xeipuuv/gojsonschema"
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeValidate, Validate)
}

type validateLogic struct {
	*model.ValidateState
	Instance
}

func Validate(instance Instance, state model.State) (Logic, error) {
	validate, ok := state.(*model.ValidateState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(validateLogic)
	sl.Instance = instance
	sl.ValidateState = validate

	return sl, nil
}

func (logic *validateLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	err := scheduleOnce(logic, wakedata)
	if err != nil {
		return nil, err
	}

	var schemaData []byte
	schemaData, err = json.Marshal(logic.Schema)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	subjectQuery := "jq(.)"
	if logic.Subject != "" {
		subjectQuery = logic.Subject
	}

	var subject interface{}
	subject, err = jqOne(logic.GetInstanceData(), subjectQuery) //nolint:contextcheck
	if err != nil {
		return nil, err
	}

	documentData, err := json.Marshal(subject)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	schema := gojsonschema.NewStringLoader(string(schemaData))
	document := gojsonschema.NewStringLoader(string(documentData))
	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	if !result.Valid() {
		for _, reason := range result.Errors() {
			logic.Log(ctx, log.Error, "Schema validation error: %s", reason.String())
		}

		return nil, derrors.NewCatchableError(ErrCodeFailedSchemaValidation, fmt.Sprintf("subject failed its JSONSchema validation: %v", err))
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
