package flow

import (
	"errors"
	"strings"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCodeInternal               = "direktiv.internal.error"
	ErrCodeWorkflowUnparsable     = "direktiv.workflow.unparsable"
	ErrCodeMultipleErrors         = "direktiv.workflow.multipleErrors"
	ErrCodeCancelledByParent      = "direktiv.cancels.parent"
	ErrCodeSoftTimeout            = "direktiv.cancels.timeout.soft"
	ErrCodeHardTimeout            = "direktiv.cancels.timeout.hard"
	ErrCodeJQBadQuery             = "direktiv.jq.badCommand"
	ErrCodeJQNotObject            = "direktiv.jq.notObject"
	ErrCodeJQNoResults            = "direktiv.jq.badCommand"
	ErrCodeJQManyResults          = "direktiv.jq.badCommand"
	ErrCodeAllBranchesFailed      = "direktiv.parallel.allFailed"
	ErrCodeNotArray               = "direktiv.foreach.badArray"
	ErrCodeFailedSchemaValidation = "direktiv.schema.failed"
	ErrCodeJQNotString            = "direktiv.jq.notString"
	ErrCodeInvalidVariableKey     = "direktiv.var.invalidKey"
)

var (
	ErrNotDir         = errors.New("not a directory")
	ErrNotWorkflow    = errors.New("not a workflow")
	ErrNotMirror      = errors.New("not a git mirror")
	ErrMirrorLocked   = errors.New("git mirror is locked")
	ErrMirrorUnlocked = errors.New("git mirror is not locked")
)

func translateError(err error) error {
	if derrors.IsNotFound(err) || errors.Is(err, filestore.ErrNotFound) {
		err = status.Error(codes.NotFound, strings.TrimPrefix(err.Error(), "ent: "))
		return err
	}

	if errors.Is(err, datastore.ErrInvalidRuntimeVariableName) {
		err = status.Error(codes.InvalidArgument, "invalid runtime variable name")
		return err
	}

	if errors.Is(err, datastore.ErrInvalidNamespaceName) {
		err = status.Error(codes.InvalidArgument, "invalid namespace name")
		return err
	}

	if strings.Contains(err.Error(), "already exists") ||
		errors.Is(err, filestore.ErrPathAlreadyExists) ||
		errors.Is(err, datastore.ErrDuplicatedNamespaceName) {
		err = status.Error(codes.AlreadyExists, "resource already exists")
		return err
	}

	return err
}
