package flow

import (
	"strings"

	"github.com/direktiv/direktiv/pkg/model"
)

// NOTE: old constants.
const (
	// Direktiv Headers.
	ServiceHeaderName          = "direktiv.io/name"
	ServiceHeaderNamespaceID   = "direktiv.io/namespace-id"
	ServiceHeaderNamespaceName = "direktiv.io/namespace-name"
	ServiceHeaderWorkflowID    = "direktiv.io/workflow-id"
	ServiceHeaderPath          = "direktiv.io/workflow-name"
	ServiceHeaderRevision      = "direktiv.io/revision"
	ServiceHeaderSize          = "direktiv.io/size"
	ServiceHeaderScale         = "direktiv.io/scale"
	ServiceTemplateGeneration  = "direktiv.io/templateGeneration"
	ServiceHeaderScope         = "direktiv.io/scope"

	// Serving Headers.
	ServiceKnativeHeaderName            = "serving.knative.dev/service"
	ServiceKnativeHeaderConfiguration   = "serving.knative.dev/configuration"
	ServiceKnativeHeaderGeneration      = "serving.knative.dev/configurationGeneration"
	ServiceKnativeHeaderRevision        = "serving.knative.dev/revision"
	ServiceKnativeHeaderRolloutDuration = "serving.knative.dev/rolloutDuration"
)

func SanitizeLabel(s string) string {
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/")
	s = strings.ReplaceAll(s, "_", "--")
	s = strings.ReplaceAll(s, "/", "-")

	if len(s) > 63 {
		s = s[:63]
	}

	return s
}

// headers for flow->container communication.
const (
	DirektivActionIDHeader    = "Direktiv-ActionID"
	DirektivInstanceIDHeader  = "Direktiv-InstanceID"
	DirektivExchangeKeyHeader = "Direktiv-ExchangeKey"
	DirektivPingAddrHeader    = "Direktiv-PingAddr"
	DirektivDeadlineHeader    = "Direktiv-Deadline"
	DirektivTimeoutHeader     = "Direktiv-Timeout"
	DirektivStepHeader        = "Direktiv-Step"
	DirektivIteratorHeader    = "Direktiv-Iterator"
	DirektivResponseHeader    = "Direktiv-Response"
	DirektivNamespaceHeader   = "Direktiv-Namespace"
	DirektivSourceHeader      = "Direktiv-Source"
	DirektivFileHeader        = "Direktiv-Files"

	DirektivErrorCodeHeader    = "Direktiv-ErrorCode"
	DirektivErrorMessageHeader = "Direktiv-ErrorMessage"
)

// internal error codes for knative services.
const (
	ServiceResponseNoError = ""
	ServiceErrorInternal   = "au.com.direktiv.error.internal"
	ServiceErrorImage      = "au.com.direktiv.error.image"
	ServiceErrorNetwork    = "au.com.direktiv.error.network"
	ServiceErrorIO         = "au.com.direktiv.error.io"
)

// ServiceResponse is the response structure for internal knative services.
type ServiceResponse struct {
	ErrorCode    string      `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	Data         interface{} `json:"data"`
}

type functionRequest struct {
	ActionID  string
	Iterator  int
	Workflow  functionWorkflow
	Container functionContainer
}

type functionContainer struct {
	Type                model.FunctionType
	ID                  string
	Image, Cmd, Service string
	Data                []byte
	Size                model.Size
	Scale               int
	Files               []model.FunctionFileDefinition
}

type functionWorkflow struct {
	Name          string
	Path          string
	InstanceID    string
	NamespaceID   string
	NamespaceName string
	State         string
	Step          int
	Timeout       int
}
