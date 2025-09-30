package core

const (
	EngineMappingPath      = "path"
	EngineMappingNamespace = "namespace"

	// do not change the values. it is used by all old containers.
	EngineHeaderActionID = "Direktiv-ActionID"
	EngineHeaderFile     = "Direktiv-File"
	// TO BE IMPLEMENTED
	// EngineHeaderTempDir  = "Direktiv-TempDir"
	// EngineHeaderErrorCode    = "Direktiv-ErrorCode"
	// EngineHeaderErrorMessage = "Direktiv-ErrorMessage".
)

type ActionPayload struct {
	Files []string `json:"files"`
	Data  any      `json:"data"`
}
