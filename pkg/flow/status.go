package flow

import (
	"context"
	"errors"
	"fmt"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/util"
)

func (engine *engine) SetInstanceFailed(ctx context.Context, im *instanceMemory, err error) error {
	var status, code, message string
	status = util.InstanceStatusFailed
	code = ErrCodeInternal

	uerr := new(derrors.UncatchableError)
	cerr := new(derrors.CatchableError)
	ierr := new(derrors.InternalError)
	engine.loggerBeta.Log(addTraceFrom(ctx, im.cached.GetAttributes("namespace")), logengine.Error, "Workflow %s canceled due to instance %s failed", im.cached.Instance.As, im.GetInstanceID())
	engine.loggerBeta.Log(addTraceFrom(ctx, im.GetAttributes()), logengine.Error, "Workflow %s canceled due to instance %s failed", im.cached.Instance.As, im.GetInstanceID())

	if errors.As(err, &uerr) {
		code = uerr.Code
		message = uerr.Message
	} else if errors.As(err, &cerr) {
		code = cerr.Code
		message = cerr.Message
	} else if errors.As(err, &ierr) {
		engine.sugar.Error(fmt.Errorf("internal error: %w", ierr))
		status = util.InstanceStatusCrashed
		message = "an internal error occurred"
	} else {
		engine.sugar.Error(fmt.Errorf("unhandled error: %w", err))
		code = ErrCodeInternal
		message = err.Error()
	}

	updater := im.getInstanceUpdater()
	updater = updater.SetStatus(status).SetErrorCode(code).SetErrorMessage(message)
	im.cached.Instance.Status = status
	im.cached.Instance.ErrorCode = code
	im.cached.Instance.ErrorMessage = message
	im.instanceUpdater = updater
	return nil
}

func (engine *engine) InstanceRaise(ctx context.Context, im *instanceMemory, cerr *derrors.CatchableError) error {
	if im.ErrorCode() == "" {
		updater := im.getInstanceUpdater()
		updater = updater.SetStatus(util.InstanceStatusFailed).SetErrorCode(cerr.Code).SetErrorMessage(cerr.Message)
		im.cached.Instance.Status = util.InstanceStatusFailed
		im.cached.Instance.ErrorCode = cerr.Code
		im.cached.Instance.ErrorMessage = cerr.Message
		im.instanceUpdater = updater
	} else {
		return derrors.NewCatchableError(ErrCodeMultipleErrors, "the workflow instance tried to throw multiple errors")
	}

	return nil
}
