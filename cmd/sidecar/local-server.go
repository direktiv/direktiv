package sidecar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/gorilla/mux"
)

const (
	workerThreads = 10
)

type LocalServer struct {
	end       func()
	flow      grpc.InternalClient
	flowAddr  string
	flowToken string
	queue     chan *inboundRequest
	router    *mux.Router
	stopper   chan *time.Time
	server    http.Server
	workers   []*inboundWorker

	requestsLock sync.Mutex
	requests     map[string]*activeRequest
}

func (srv *LocalServer) initFlow() error {
	serverArr := fmt.Sprintf("%s:7777", os.Getenv(direktivFlowEndpoint))
	fmt.Printf("flow server: %s\n", serverArr)
	conn, err := utils.GetEndpointTLS(serverArr)
	if err != nil {
		return err
	}
	fmt.Printf("connected to flow\n")

	srv.flow = grpc.NewInternalClient(conn)

	srv.flowToken = os.Getenv("API_KEY")
	srv.flowAddr = fmt.Sprintf("%s:6665", os.Getenv(direktivFlowEndpoint))

	return nil
}

func (srv *LocalServer) Start() {
	err := srv.initFlow()
	if err != nil {
		slog.Error("Localhost server unable to connect to flow", "error", err)
		Shutdown(ERROR)

		return
	}

	srv.queue = make(chan *inboundRequest, 100)
	srv.requests = make(map[string]*activeRequest)

	srv.router = mux.NewRouter()

	srv.router.Use(utils.TelemetryMiddleware)

	srv.router.HandleFunc("/log", srv.logHandler)
	srv.router.HandleFunc("/var", srv.varHandler)

	srv.server.Addr = "127.0.0.1:8889"
	srv.server.Handler = srv.router

	srv.stopper = make(chan *time.Time, 1)

	srv.end = threads.Register(srv.stopper)

	slog.Debug("Localhost server thread registered.")

	//nolint:intrange
	for i := 0; i < workerThreads; i++ {
		worker := new(inboundWorker)
		worker.id = i
		worker.srv = srv
		srv.workers = append(srv.workers, worker)
		go worker.run()
	}

	go srv.run()
	go srv.wait()
}

func (srv *LocalServer) wait() {
	defer srv.server.Close()
	defer srv.end()

	t := <-srv.stopper
	close(srv.queue)

	slog.Debug("Localhost server shutting down.")

	for req := range srv.queue {
		go srv.drainRequest(req)
	}

	for _, worker := range srv.workers {
		go worker.Cancel()
	}

	ctx, cancel := context.WithDeadline(context.Background(), t.Add(20*time.Second))
	defer cancel()

	err := srv.server.Shutdown(ctx)
	if err != nil {
		slog.Error("Error shutting down localhost server", "error", err)
		Shutdown(ERROR)

		return
	}

	slog.Debug("Primary localhost server thread shut down successfully.")
}

func (srv *LocalServer) logHandler(w http.ResponseWriter, r *http.Request) {
	actionId := r.URL.Query().Get("aid")

	srv.requestsLock.Lock()
	req, ok := srv.requests[actionId]
	srv.requestsLock.Unlock()

	reportError := func(code int, err error) {
		http.Error(w, err.Error(), code)
		slog.Warn("Log handler error occurred.", "action_id", actionId, "action_err_code", code, "error", err)
	}

	if !ok {
		err := errors.New("the action id went missing")
		code := http.StatusInternalServerError
		reportError(code, err)

		return
	}

	if req == nil {
		code := http.StatusNotFound
		reportError(code, fmt.Errorf("actionId %s not found", actionId))

		return
	}

	var msg string

	if r.Method == http.MethodPost {
		capa := int64(0x400000) // 4 MiB
		if r.ContentLength > capa {
			code := http.StatusRequestEntityTooLarge
			reportError(code, errors.New(http.StatusText(code)))

			return
		}
		r := io.LimitReader(r.Body, capa)

		data, err := io.ReadAll(r)
		if err != nil {
			code := http.StatusBadRequest
			reportError(code, err)

			return
		}

		msg = string(data)
	} else {
		msg = r.URL.Query().Get("log")
	}

	if len(msg) == 0 {
		slog.Debug("Log handler received an empty message body.", "action_id", actionId)
		return
	}

	_, err := srv.flow.ActionLog(req.ctx, &grpc.ActionLogRequest{
		InstanceId: req.instanceId,
		Msg:        []string{msg},
		Iterator:   int32(req.iterator),
	})
	if err != nil {
		slog.Error("Failed to forward log to Flow.", "action_id", actionId, "error", err)
	}

	slog.Debug("Log handler successfully processed message.", "action_id", actionId)
}

// nolint:canonicalheader
func (srv *LocalServer) varHandler(w http.ResponseWriter, r *http.Request) {
	actionId := r.URL.Query().Get("aid")

	srv.requestsLock.Lock()
	req, ok := srv.requests[actionId]
	srv.requestsLock.Unlock()

	reportError := func(code int, err error) {
		http.Error(w, err.Error(), code)
		slog.Warn("Variable retrieval failed.", "action_id", actionId, "error", err)
	}

	if !ok {
		err := errors.New("the action id went missing")
		code := http.StatusInternalServerError
		reportError(code, err)

		return
	}

	if req == nil {
		code := http.StatusNotFound
		reportError(code, fmt.Errorf("actionId %s not found", actionId))

		return
	}

	ctx := req.ctx
	ir := req.functionRequest

	scope := r.URL.Query().Get("scope")
	key := r.URL.Query().Get("key")
	vMimeType := r.Header.Get("content-type")

	switch r.Method {
	case http.MethodGet:

		varMeta, statusCode, err := getVariableMetaFromFlow(ctx, srv.flowToken, srv.flowAddr, ir, scope, key)
		if err != nil {
			reportError(statusCode, err)
			slog.Warn("Failed retrieving a Variable.", "action_id", actionId, "key", key, "scope", scope)

			return
		}

		varData, err := getVariableDataViaID(ctx, srv.flowToken, srv.flowAddr, ir.namespace, varMeta.ID.String())
		if err != nil {
			reportError(http.StatusInternalServerError, err)
			slog.Warn("Failed retrieving a Variable.", "action_id", actionId, "key", key, "scope", scope)

			return
		}
		_, err = io.Copy(w, bytes.NewReader(varData.Data))
		if err != nil {
			reportError(http.StatusInternalServerError, err)
			slog.Error("Failed retrieving a Variable.", "action_id", actionId, "key", key, "scope", scope)

			return
		}

		slog.Debug("Variable successfully retrieved.", "action_id", actionId, "key", key, "scope", scope)

	case http.MethodPost:

		statusCode, err := srv.setVar(ctx, ir, r.Body, scope, key, vMimeType)
		if err != nil {
			reportError(statusCode, err)
			slog.Warn("Failed to set a Variable.", "action_id", actionId, "key", key, "scope", scope)

			return
		}

		slog.Debug("Variable successfully stored.", "action_id", actionId, "key", key, "scope", scope, "mime_type", vMimeType)

	default:
		code := http.StatusMethodNotAllowed
		reportError(code, errors.New(http.StatusText(code)))
		slog.Warn("Unsupported HTTP method for var handler.", "action_id", actionId, "method", r.Method)

		return
	}
}

type activeRequest struct {
	*functionRequest
	cancel func()
	ctx    context.Context //nolint:containedctx
}

func (srv *LocalServer) registerActiveRequest(ir *functionRequest, ctx context.Context, cancel func()) {
	srv.requestsLock.Lock()

	srv.requests[ir.actionId] = &activeRequest{
		functionRequest: ir,
		ctx:             ctx,
		cancel:          cancel,
	}

	srv.requestsLock.Unlock()

	slog.Info("Serving.", "action_id", ir.actionId)
}

func (srv *LocalServer) deregisterActiveRequest(actionId string) {
	srv.requestsLock.Lock()

	delete(srv.requests, actionId)

	srv.requestsLock.Unlock()

	slog.Debug("Request deregistered.", "action_id", actionId)
}

func (srv *LocalServer) cancelActiveRequest(ctx context.Context, actionId string) {
	srv.requestsLock.Lock()
	req := srv.requests[actionId]
	srv.requestsLock.Unlock()

	if req == nil {
		return
	}

	slog.Info("Attempting to cancel.", "action_id", actionId)

	go srv.sendCancelToService(ctx, req.functionRequest)

	select {
	case <-req.ctx.Done():
	case <-time.After(10 * time.Second):
		slog.Warn("Request failed to cancel punctually.", "action_id", actionId)
		req.cancel()
	}
}

func (srv *LocalServer) sendCancelToService(ctx context.Context, ir *functionRequest) {
	url := "http://localhost:8080"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		slog.Error("Failed to create cancel request.", "action_id", ir.actionId, "error", err)
		return
	}

	req.Header.Set(actionIDHeader, ir.actionId)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to send cancel to service.", "action_id", ir.actionId, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Service responded to cancel request.", "action_id", ir.actionId, "resp_code", resp.StatusCode)
	}
}

type inboundRequest struct {
	w   http.ResponseWriter
	r   *http.Request
	end chan bool
}

func (srv *LocalServer) drainRequest(req *inboundRequest) {
	_ = req.r.Body.Close()

	code := http.StatusServiceUnavailable
	msg := http.StatusText(code)
	http.Error(req.w, msg, code)

	id := req.r.Header.Get(actionIDHeader)
	slog.Warn("Request aborted due to server unavailability", "action_id", id, "http_status_code", code, "reason", msg)

	defer func() {
		_ = recover()
	}()

	close(req.end)
}

func (srv *LocalServer) run() {
	slog.Info("Starting localhost HTTP server.", "addr", srv.server.Addr)

	err := srv.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Error running local server", "error", err)
		Shutdown(ERROR)

		return
	}
}

type functionRequest struct {
	actionId     string
	instanceId   string
	namespace    string
	workflowPath string
	step         int
	deadline     time.Time
	input        []byte
	files        []*functionFiles
	iterator     int
}

type functionFiles struct {
	Key         string `json:"key"`
	As          string `json:"as"`
	Scope       string `json:"scope"`
	Type        string `json:"type"`
	Permissions string `json:"permissions"`
}

const sharedDir = "/mnt/shared"

func (srv *LocalServer) setVar(ctx context.Context, ir *functionRequest, r io.Reader, scope, key, vMimeType string) (int, error) {
	// Retrieve variable metadata
	varMeta, statusCode, err := getVariableMetaFromFlow(ctx, srv.flowToken, srv.flowAddr, ir, scope, key)
	if err != nil {
		target := &RessourceNotFoundError{}
		if errors.As(err, &target) {
			data, readErr := io.ReadAll(r)
			if readErr != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to read data from reader: %w", readErr)
			}

			reqD := createVarRequest{
				Name:     key,
				MimeType: vMimeType,
				Data:     data,
			}

			// Set scope-specific fields
			switch scope {
			case utils.VarScopeInstance:
				reqD.InstanceIDString = ir.instanceId
			case utils.VarScopeWorkflow:
				reqD.WorkflowPath = ir.workflowPath
			case utils.VarScopeNamespace:
				// Namespace scope requires no additional fields
			default:
				return http.StatusBadRequest, fmt.Errorf("unknown scope: %s", scope)
			}

			// Attempt to create the variable
			postStatusCode, postErr := postVarData(ctx, srv.flowToken, srv.flowAddr, ir.namespace, reqD)
			if postErr != nil {
				return postStatusCode, fmt.Errorf("failed to post variable data: %w", postErr)
			}
			return http.StatusOK, nil
		}
		// Handle other errors from getVariableMetaFromFlow
		return statusCode, fmt.Errorf("failed to get variable metadata: %w", err)
	}

	// Patch existing variable data
	data, readErr := io.ReadAll(r)
	if readErr != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to read data from reader: %w", readErr)
	}

	reqD := datastore.RuntimeVariablePatch{
		Name:     &key,
		MimeType: &vMimeType,
		Data:     data,
	}
	patchStatusCode, patchErr := patchVarData(ctx, srv.flowToken, srv.flowAddr, ir.namespace, varMeta.ID.String(), reqD)
	if patchErr != nil {
		return patchStatusCode, fmt.Errorf("failed to patch variable data: %w", patchErr)
	}
	return statusCode, nil
}
