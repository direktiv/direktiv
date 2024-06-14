package flow

import (
	"errors"
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
