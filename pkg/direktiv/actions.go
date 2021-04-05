package direktiv

import "github.com/vorteil/direktiv/pkg/model"

// headers for flow->container communication
const (
	DirektivActionIDHeader    = "Direktiv-ActionID"
	DirektivInstanceIDHeader  = "Direktiv-InstanceID"
	DirektivExchangeKeyHeader = "Direktiv-ExchangeKey"
	DirektivPingAddrHeader    = "Direktiv-PingAddr"
	DirektivTimeoutHeader     = "Direktiv-Timeout"
	DirektivStepHeader        = "Direktiv-Step"
	DirektivResponseHeader    = "Direktiv-Response"

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

type actionRequest struct {
	ActionID string

	Workflow  actionWorkflow
	Container actionContainer
}

type actionContainer struct {
	Image, Cmd string
	Data       []byte
	Size       model.Size
}

type actionWorkflow struct {
	Name       string
	InstanceID string
	Namespace  string
	State      string
	Step       int
	Timeout    int
}
