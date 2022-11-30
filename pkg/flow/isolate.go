package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
)

// headers for flow->container communication.
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
	ActionID string

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
	WorkflowID    string
	Revision      string
	InstanceID    string
	NamespaceID   string
	NamespaceName string
	State         string
	Step          int
	Timeout       int
}

func (engine *engine) isScopedKnativeFunction(client igrpc.FunctionsServiceClient,
	serviceName string) bool {

	// search annotations
	a := make(map[string]string)
	a[functions.ServiceKnativeHeaderName] = serviceName

	engine.sugar.Debugf("knative function search: %v", a)

	_, err := client.GetFunction(context.Background(), &igrpc.GetFunctionRequest{
		ServiceName: &serviceName,
	})

	if err != nil {
		engine.sugar.Errorf("can not get knative service: %v", err)
		return false
	}

	return true
}

func reconstructScopedKnativeFunction(client igrpc.FunctionsServiceClient,
	serviceName string) error {

	cr := igrpc.ReconstructFunctionRequest{
		Name: &serviceName,
	}

	_, err := client.ReconstructFunction(context.Background(), &cr)
	return err
}

func (engine *engine) isKnativeFunction(client igrpc.FunctionsServiceClient, ar *functionRequest) bool {

	// search annotations
	a := make(map[string]string)
	a[functions.ServiceHeaderName] = functions.SanitizeLabel(ar.Container.ID)
	a[functions.ServiceHeaderNamespaceID] = functions.SanitizeLabel(ar.Workflow.NamespaceID)
	a[functions.ServiceHeaderWorkflowID] = functions.SanitizeLabel(ar.Workflow.WorkflowID)

	engine.sugar.Debugf("knative function search: %v", a)

	l, err := client.ListFunctions(context.Background(), &igrpc.ListFunctionsRequest{
		Annotations: a,
	})

	if err != nil {
		engine.sugar.Errorf("can not list knative service: %v", err)
		return false
	}

	if len(l.Functions) > 0 {
		return true
	}

	return false

}

func createKnativeFunction(client igrpc.FunctionsServiceClient,
	ir *functionRequest) error {

	sz := int32(ir.Container.Size)
	scale := int32(ir.Container.Scale)

	cr := igrpc.CreateFunctionRequest{
		Info: &igrpc.BaseInfo{
			Name:          &ir.Container.ID,
			Namespace:     &ir.Workflow.NamespaceID,
			Workflow:      &ir.Workflow.WorkflowID,
			Image:         &ir.Container.Image,
			Cmd:           &ir.Container.Cmd,
			Size:          &sz,
			MinScale:      &scale,
			NamespaceName: &ir.Workflow.NamespaceName,
			Path:          &ir.Workflow.Path,
			Revision:      &ir.Workflow.Revision,
		},
	}

	_, err := client.CreateFunction(context.Background(), &cr)

	return err

}
