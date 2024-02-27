package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
)

func (engine *engine) GetIsInstanceFailed(im *instanceMemory) bool {
	if engine.GetIsInstanceCrashed(im) {
		return true
	}

	if im.instance.Instance.Status == instancestore.InstanceStatusFailed {
		return true
	}

	if im.updateArgs.Status != nil && (*im.updateArgs.Status) == instancestore.InstanceStatusFailed {
		return true
	}

	return false
}

func (engine *engine) GetIsInstanceCrashed(im *instanceMemory) bool {
	if im.instance.Instance.Status == instancestore.InstanceStatusCrashed {
		return true
	}

	if im.updateArgs.Status != nil && (*im.updateArgs.Status) == instancestore.InstanceStatusCrashed {
		return true
	}

	return false
}

func (engine *engine) SetInstanceFailed(ctx context.Context, im *instanceMemory, err error) {
	var status instancestore.InstanceStatus
	var code, message string
	status = instancestore.InstanceStatusFailed
	code = ErrCodeInternal

	uerr := new(derrors.UncatchableError)
	cerr := new(derrors.CatchableError)
	ierr := new(derrors.InternalError)
	engine.logger.Errorf(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "Workflow %s canceled due to instance %s failed", im.instance.Instance.WorkflowPath, im.GetInstanceID())
	slog.Error(fmt.Sprintf("Workflow %s canceled due to instance %s failed", im.instance.Instance.WorkflowPath, im.GetInstanceID()), "stream", recipient.Namespace.String()+"."+im.Namespace().Name)
	engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), "Workflow %s canceled due to instance %s failed", im.instance.Instance.WorkflowPath, im.GetInstanceID())
	slog.Error(fmt.Sprintf("Workflow %s canceled due to instance %s failed", im.instance.Instance.WorkflowPath, im.GetInstanceID()), im.GetSlogAttributes(ctx)...)
	if errors.As(err, &uerr) {
		code = uerr.Code
		message = uerr.Message
	} else if errors.As(err, &cerr) {
		code = cerr.Code
		message = cerr.Message
	} else if errors.As(err, &ierr) {
		engine.sugar.Error(fmt.Errorf("internal error: %w", ierr))
		status = instancestore.InstanceStatusCrashed
		message = "an internal error occurred"
	} else {
		engine.sugar.Error(fmt.Errorf("unhandled error: %w", err))
		status = instancestore.InstanceStatusCrashed
		code = ErrCodeInternal
		message = err.Error()
	}

	im.instance.Instance.Status = status
	im.instance.Instance.ErrorCode = code
	im.instance.Instance.ErrorMessage = []byte(message)
	im.updateArgs.Status = &im.instance.Instance.Status
	im.updateArgs.ErrorCode = &im.instance.Instance.ErrorCode
	im.updateArgs.ErrorMessage = &im.instance.Instance.ErrorMessage
}

func (engine *engine) InstanceRaise(ctx context.Context, im *instanceMemory, cerr *derrors.CatchableError) error {
	if im.ErrorCode() == "" {
		im.instance.Instance.Status = instancestore.InstanceStatusFailed
		im.instance.Instance.ErrorCode = cerr.Code
		im.instance.Instance.ErrorMessage = []byte(cerr.Message)
		im.updateArgs.Status = &im.instance.Instance.Status
		im.updateArgs.ErrorCode = &im.instance.Instance.ErrorCode
		im.updateArgs.ErrorMessage = &im.instance.Instance.ErrorMessage
	} else {
		return derrors.NewCatchableError(ErrCodeMultipleErrors, "the workflow instance tried to throw multiple errors")
	}

	return nil
}
