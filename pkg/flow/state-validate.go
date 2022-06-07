package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
	"github.com/xeipuuv/gojsonschema"
)

type validateStateLogic struct {
	*model.ValidateState
}

func initValidateStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	validate, ok := state.(*model.ValidateState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(validateStateLogic)
	sl.ValidateState = validate

	return sl, nil
}

func (sl *validateStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
}

func (sl *validateStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *validateStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	var schemaData []byte
	schemaData, err = json.Marshal(sl.Schema)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	subjectQuery := "jq(.)"
	if sl.Subject != "" {
		subjectQuery = sl.Subject
	}

	var subject interface{}
	subject, err = jqObject(im.data, subjectQuery)
	if err != nil {
		return
	}

	documentData, err := json.Marshal(subject)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	schema := gojsonschema.NewStringLoader(string(schemaData))
	document := gojsonschema.NewStringLoader(string(documentData))
	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		err = NewInternalError(err)
		return
	}

	if !result.Valid() {
		for _, reason := range result.Errors() {
			engine.logToInstance(ctx, time.Now(), im.in, "Schema validation error: %s", reason.String())
		}
		err = NewCatchableError(ErrCodeFailedSchemaValidation, fmt.Sprintf("subject failed its JSONSchema validation: %v", err))
		return
	}

	transition = &stateTransition{
		Transform: sl.Transform,
		NextState: sl.Transition,
	}

	return

}
