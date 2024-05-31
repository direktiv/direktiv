package runtime

const (
	InstancesDir = "instances"
	SharedDir    = "shared"
)

const (
	DirektivActionIDHeader = "Direktiv-ActionID"
	DirektivTempDir        = "Direktiv-TempDir"

	DirektivErrorCodeHeader    = "Direktiv-ErrorCode"
	DirektivErrorMessageHeader = "Direktiv-ErrorMessage"

	DirektivErrorCode = "io.direktiv.error.execution"

	DirektivFileErrorCode     = "io.direktiv.error.file"
	DirektivSecretsErrorCode  = "io.direktiv.error.secrets"
	DirektivHTTPErrorCode     = "io.direktiv.error.http"
	DirektivTimeoutErrorCode  = "io.direktiv.error.timeout"
	DirektivFunctionErrorCode = "io.direktiv.error.function"
	DirektivErrorInternal     = "io.direktiv.internal"
)

const StateDataInputFile = "input.data"
