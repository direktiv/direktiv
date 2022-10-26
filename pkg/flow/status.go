package flow

import (
	"context"
	"fmt"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/util"
)

func (engine *engine) SetInstanceFailed(ctx context.Context, im *instanceMemory, err error) error {

	var status, code, message string
	status = util.InstanceStatusFailed
	code = ErrCodeInternal

	if uerr, ok := err.(*derrors.UncatchableError); ok {
		code = uerr.Code
		message = uerr.Message
	} else if cerr, ok := err.(*derrors.CatchableError); ok {
		code = cerr.Code
		message = cerr.Message
	} else if _, ok := err.(*derrors.InternalError); ok {
		engine.sugar.Error(fmt.Errorf("internal error: %v", err))
		status = util.InstanceStatusCrashed
		message = "an internal error occurred"
	} else {
		engine.sugar.Error(fmt.Errorf("Unhandled error: %v", err))
		code = ErrCodeInternal
		message = err.Error()
	}

	updater := im.getInstanceUpdater()
	updater = updater.SetStatus(status).SetErrorCode(code).SetErrorMessage(message)
	im.in.Status = status
	im.in.ErrorCode = code
	im.in.ErrorMessage = message
	im.instanceUpdater = updater

	return nil

}

func (engine *engine) InstanceRaise(ctx context.Context, im *instanceMemory, cerr *derrors.CatchableError) error {

	if im.ErrorCode() == "" {

		updater := im.getInstanceUpdater()
		updater = updater.SetStatus(util.InstanceStatusFailed).SetErrorCode(cerr.Code).SetErrorMessage(cerr.Message)
		im.in.Status = util.InstanceStatusFailed
		im.in.ErrorCode = cerr.Code
		im.in.ErrorMessage = cerr.Message
		im.instanceUpdater = updater

	} else {
		return derrors.NewCatchableError(ErrCodeMultipleErrors, "the workflow instance tried to throw multiple errors")
	}

	return nil

}
