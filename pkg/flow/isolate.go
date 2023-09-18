package flow

import (
	"context"
	"net/http"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
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
	Headers   http.Header
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
	Revision      string
	InstanceID    string
	NamespaceID   string
	NamespaceName string
	State         string
	Step          int
	Timeout       int
}

func (engine *engine) isScopedKnativeFunction(functionsClient igrpc.FunctionsClient,
	serviceName string,
) bool {
	// search annotations
	a := make(map[string]string)
	a[functions.ServiceKnativeHeaderName] = serviceName

	engine.sugar.Debugf("knative function search: %v", a)

	_, err := functionsClient.GetFunction(context.Background(), &igrpc.FunctionsGetFunctionRequest{
		ServiceName: &serviceName,
	})
	if err != nil {
		engine.sugar.Errorf("can not get knative service: %v", err)
		return false
	}

	return true
}

func reconstructScopedKnativeFunction(functionsClient igrpc.FunctionsClient,
	serviceName string,
) error {
	cr := igrpc.FunctionsReconstructFunctionRequest{
		Name: &serviceName,
	}

	_, err := functionsClient.ReconstructFunction(context.Background(), &cr)
	return err
}

func (engine *engine) isKnativeFunction(functionsClient igrpc.FunctionsClient, ar *functionRequest) bool {
	// search annotations
	a := make(map[string]string)
	a[functions.ServiceHeaderName] = functions.SanitizeLabel(ar.Container.ID)
	a[functions.ServiceHeaderNamespaceID] = functions.SanitizeLabel(ar.Workflow.NamespaceID)
	a[functions.ServiceHeaderWorkflowID] = functions.SanitizeLabel(bytedata.ShortChecksum(ar.Workflow.Path))

	engine.sugar.Debugf("knative function search: %v", a)

	l, err := functionsClient.ListFunctions(context.Background(), &igrpc.FunctionsListFunctionsRequest{
		Annotations: a,
	})
	if err != nil {
		engine.sugar.Errorf("can not list knative service: %v", err)
		return false
	}

	if len(l.Functions) > 0 {
		engine.sugar.Debugf("found functions")
		return true
	}
	engine.sugar.Debugf("no functions found")

	return false
}

func createKnativeFunction(functionsClient igrpc.FunctionsClient,
	ir *functionRequest,
) error {
	sz := int32(ir.Container.Size)
	scale := int32(ir.Container.Scale)

	wf := bytedata.ShortChecksum(ir.Workflow.Path)

	cr := igrpc.FunctionsCreateFunctionRequest{
		Info: &igrpc.FunctionsBaseInfo{
			Name:          &ir.Container.ID,
			Namespace:     &ir.Workflow.NamespaceID,
			Image:         &ir.Container.Image,
			Cmd:           &ir.Container.Cmd,
			Size:          &sz,
			MinScale:      &scale,
			NamespaceName: &ir.Workflow.NamespaceName,
			Path:          &ir.Workflow.Path,
			Revision:      &ir.Workflow.Revision,
			Workflow:      &wf,
		},
	}

	_, err := functionsClient.CreateFunction(context.Background(), &cr)

	return err
}
