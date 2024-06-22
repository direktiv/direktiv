package flow

var (
	ErrCodeInternal           = "direktiv.internal.error"
	ErrCodeWorkflowUnparsable = "direktiv.workflow.unparsable"
	ErrCodeMultipleErrors     = "direktiv.workflow.multipleErrors"
	ErrCodeCancelledByParent  = "direktiv.cancels.parent"
	ErrCodeSoftTimeout        = "direktiv.cancels.timeout.soft"
	ErrCodeHardTimeout        = "direktiv.cancels.timeout.hard"
	ErrCodeJQBadQuery         = "direktiv.jq.badCommand"
	ErrCodeJQNotObject        = "direktiv.jq.notObject"
	ErrCodeJQNoResults        = "direktiv.jq.badCommand"
	ErrCodeJQManyResults      = "direktiv.jq.badCommand"
)
