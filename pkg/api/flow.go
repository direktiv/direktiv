package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
)

type flowHandler struct {
	logger *zap.SugaredLogger
	client grpc.FlowClient
}

func newFlowHandler(logger *zap.SugaredLogger, router *mux.Router, addr string) (*flowHandler, error) {

	flowAddr := fmt.Sprintf("%s:6666", addr)
	logger.Infof("connecting to flow %s", flowAddr)

	conn, err := util.GetEndpointTLS(flowAddr)
	if err != nil {
		logger.Errorf("can not connect to direktiv flows: %v", err)
		return nil, err
	}

	h := &flowHandler{
		logger: logger,
		client: grpc.NewFlowClient(conn),
	}

	h.initRoutes(router)

	return h, nil

}

func (h *flowHandler) initRoutes(r *mux.Router) {
	r.HandleFunc("/namespaces", h.Namespace)
}

func (h *flowHandler) Namespace(w http.ResponseWriter, r *http.Request) {

	// name := "" // TODO

}

func (h *flowHandler) listFunctions(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("LIST FLOW")
	w.Write([]byte("LIST FLOW"))
}
