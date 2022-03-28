package flow

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/util"
)

func (engine *engine) SetInstanceFailed(ctx context.Context, im *instanceMemory, err error) error {

	var status, code, message string
	status = util.InstanceStatusFailed
	code = ErrCodeInternal

	if uerr, ok := err.(*UncatchableError); ok {
		code = uerr.Code
		message = uerr.Message
	} else if cerr, ok := err.(*CatchableError); ok {
		code = cerr.Code
		message = cerr.Message
	} else if _, ok := err.(*InternalError); ok {
		engine.sugar.Error(fmt.Errorf("internal error: %v", err))
		status = util.InstanceStatusCrashed
		message = "an internal error occurred"
	} else {
		engine.sugar.Error(fmt.Errorf("Unhandled error: %v", err))
		return nil
	}

	//.SetEndTime
	in, err := im.in.Update().SetStatus(status).SetErrorCode(code).SetErrorMessage(message).Save(ctx)
	if err != nil {
		return NewInternalError(err)
	}
	in.Edges = im.in.Edges
	im.in = in

	engine.pubsub.NotifyInstance(im.in)
	if ns, err := im.in.Namespace(ctx); err == nil {
		engine.pubsub.NotifyInstances(ns)
	}

	return nil

}

func (engine *engine) InstanceRaise(ctx context.Context, im *instanceMemory, cerr *CatchableError) error {

	if im.ErrorCode() == "" {

		in, err := im.in.Update().SetStatus(util.InstanceStatusFailed).SetErrorCode(cerr.Code).SetErrorMessage(cerr.Message).Save(ctx)
		if err != nil {
			return NewInternalError(err)
		}

		in.Edges = im.in.Edges
		im.in = in

	} else {
		return NewCatchableError(ErrCodeMultipleErrors, "the workflow instance tried to throw multiple errors")
	}

	return nil

}
