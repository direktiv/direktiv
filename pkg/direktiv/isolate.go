package direktiv

import "github.com/vorteil/direktiv/pkg/model"

// headers for flow->container communication
const (
	DirektivActionIDHeader    = "Direktiv-ActionID"
	DirektivInstanceIDHeader  = "Direktiv-InstanceID"
	DirektivExchangeKeyHeader = "Direktiv-ExchangeKey"
	DirektivPingAddrHeader    = "Direktiv-PingAddr"
	DirektivDeadlineHeader    = "Direktiv-Deadline"
	DirektivTimeoutHeader     = "Direktiv-Timeout"
	DirektivStepHeader        = "Direktiv-Step"
	DirektivResponseHeader    = "Direktiv-Response"
	DirektivNamespaceHeader   = "Direktiv-Namespace"
	DirektivSourceHeader      = "Direktiv-Source"
	DirektivFileHeader        = "Direktiv-Files"

	DirektivErrorCodeHeader    = "Direktiv-ErrorCode"
	DirektivErrorMessageHeader = "Direktiv-ErrorMessage"
)

// internal error codes for knative services
const (
	ServiceResponseNoError = ""
	ServiceErrorInternal   = "au.com.direktiv.error.internal"
	ServiceErrorImage      = "au.com.direktiv.error.image"
	ServiceErrorNetwork    = "au.com.direktiv.error.network"
	ServiceErrorIO         = "au.com.direktiv.error.io"
)

// ServiceResponse is the response structure for internal knative services
type ServiceResponse struct {
	ErrorCode    string      `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	Data         interface{} `json:"data"`
}

type isolateRequest struct {
	ActionID string

	Workflow  isolateWorkflow
	Container isolateContainer
}

type isolateContainer struct {
	ID         string
	Image, Cmd string
	Data       []byte
	Size       model.Size
	Scale      int
	Files      []model.FunctionFileDefinition
}

type isolateWorkflow struct {
	Name       string
	ID         string
	InstanceID string
	Namespace  string
	State      string
	Step       int
	Timeout    int
}
