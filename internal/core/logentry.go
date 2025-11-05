package core

type LogStatus string

const (
	LogErrStatus       LogStatus = "error"
	LogUnknownStatus   LogStatus = "unknown"
	LogRunningStatus   LogStatus = "running"
	LogFailedStatus    LogStatus = "failed"
	LogCompletedStatus LogStatus = "completed"
)
