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
	"strconv"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/gorilla/mux"
)

const (
	workerThreads = 10
)

type LocalServer struct {
	end     func()
	flow    grpc.InternalClient
	queue   chan *inboundRequest
	router  *mux.Router
	stopper chan *time.Time
	server  http.Server
	workers []*inboundWorker

	requestsLock sync.Mutex
	requests     map[string]*activeRequest
}

func (srv *LocalServer) initFlow() error {
	serverArr := fmt.Sprintf("%s:7777", os.Getenv(direktivFlowEndpoint))
	fmt.Printf("flow server: %s\n", serverArr)
	conn, err := util.GetEndpointTLS(serverArr)
	if err != nil {
		return err
	}
	fmt.Printf("connected to flow\n")

	srv.flow = grpc.NewInternalClient(conn)

	return nil
}

func (srv *LocalServer) Start() {
	err := srv.initFlow()
	if err != nil {
		log.Errorf("Localhost server unable to connect to flow: %v", err)
		Shutdown(ERROR)
		return
	}

	srv.queue = make(chan *inboundRequest, 100)
	srv.requests = make(map[string]*activeRequest)

	srv.router = mux.NewRouter()

	srv.router.Use(util.TelemetryMiddleware)

	srv.router.HandleFunc("/log", srv.logHandler)
	srv.router.HandleFunc("/var", srv.varHandler)

	srv.server.Addr = "127.0.0.1:8889"
	srv.server.Handler = srv.router

	srv.stopper = make(chan *time.Time, 1)

	srv.end = threads.Register(srv.stopper)

	log.Debug("Localhost server thread registered.")

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

	log.Debug("Localhost server shutting down.")

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
		log.Errorf("Error shutting down localhost server: %v", err)
		Shutdown(ERROR)
		return
	}

	log.Debug("Primary localhost server thread shut down successfully.")
}

func (srv *LocalServer) logHandler(w http.ResponseWriter, r *http.Request) {
	actionId := r.URL.Query().Get("aid")

	srv.requestsLock.Lock()
	req, ok := srv.requests[actionId]
	srv.requestsLock.Unlock()

	reportError := func(code int, err error) {
		http.Error(w, err.Error(), code)
		log.Warnf("Log handler for '%s' returned %v: %v.", actionId, code, err)
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
		cap := int64(0x400000) // 4 MiB
		if r.ContentLength > cap {
			code := http.StatusRequestEntityTooLarge
			reportError(code, errors.New(http.StatusText(code)))
			return
		}
		r := io.LimitReader(r.Body, cap)

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
		log.Debugf("Log handler for '%s' received zero bytes.", actionId)
		return
	}
	if ok {
		root, _ := req.fnCtx.Callers.RootInstanceID()
		if root == "" {
			root = req.fnCtx.InstanceID.String()
		}

		slog.Info(msg,
			"stream", "instance."+root,
			"workflow", req.fnCtx.WorkflowPath,
			"callpath", req.fnCtx.Callers.Callpath(),
			"root-instance-id", root,
			"state", req.fnCtx.State,
			"branch", req.fnCtx.Branch,
			"trace", req.fnCtx.Info.InstanceTelemetryInfo.TraceID,
			"async", req.fnCtx.AsyncExec,
		)
	}
	_, err := srv.flow.ActionLog(req.ctx, &grpc.ActionLogRequest{
		InstanceId: req.instanceId,
		Msg:        []string{msg},
		Iterator:   int32(req.iterator),
	})
	if err != nil {
		log.Errorf("Failed to forward log to diretiv: %v.", err)
	}

	log.Debugf("Log handler for '%s' posted %d bytes.", actionId, len(msg))
}

func (srv *LocalServer) varHandler(w http.ResponseWriter, r *http.Request) {
	actionId := r.URL.Query().Get("aid")

	srv.requestsLock.Lock()
	req, ok := srv.requests[actionId]
	srv.requestsLock.Unlock()

	reportError := func(code int, err error) {
		http.Error(w, err.Error(), code)
		log.Warnf("Var handler for '%s' returned %v: %v.", actionId, code, err)
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

		setTotalSize := func(x int64) {
			w.Header().Set("Content-Length", strconv.Itoa(int(x)))
		}

		err := srv.getVar(ctx, ir, w, setTotalSize, scope, key)
		if err != nil {
			reportError(http.StatusInternalServerError, err)
			return
		}

		log.Debugf("Var handler for '%s' retrieved %s (%s)", actionId, key, scope)

	case http.MethodPost:

		err := srv.setVar(ctx, ir, r.ContentLength, r.Body, scope, key, vMimeType)
		if err != nil {
			reportError(http.StatusInternalServerError, err)
			return
		}

		log.Debugf("Var handler for '%s' stored %s (%s)", actionId, key, scope)

	default:
		code := http.StatusMethodNotAllowed
		reportError(code, errors.New(http.StatusText(code)))
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

	log.Infof("Serving '%s'.", ir.actionId)
}

func (srv *LocalServer) deregisterActiveRequest(actionId string) {
	srv.requestsLock.Lock()

	delete(srv.requests, actionId)

	srv.requestsLock.Unlock()

	log.Debugf("Request deregistered '%s'.", actionId)
}

func (srv *LocalServer) cancelActiveRequest(ctx context.Context, actionId string) {
	srv.requestsLock.Lock()
	req := srv.requests[actionId]
	srv.requestsLock.Unlock()

	if req == nil {
		return
	}

	log.Infof("Attempting to cancel '%s'.", actionId)

	go srv.sendCancelToService(ctx, req.functionRequest)

	select {
	case <-req.ctx.Done():
	case <-time.After(10 * time.Second):
		log.Warnf("Request '%s' failed to cancel punctually.", actionId)
		req.cancel()
	}
}

func (srv *LocalServer) sendCancelToService(ctx context.Context, ir *functionRequest) {
	url := "http://localhost:8080"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.Errorf("Failed to create cancel request for '%s': %v.", ir.actionId, err)
		return
	}

	req.Header.Set(actionIDHeader, ir.actionId)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to send cancel to service for '%s': %v.", ir.actionId, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("Service responded to cancel request for '%s' with %v.", ir.actionId, resp.StatusCode)
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
	log.Warnf("Aborting request '%s' early.", id)

	defer func() {
		_ = recover()
	}()

	close(req.end)
}

func (srv *LocalServer) run() {
	log.Infof("Starting localhost HTTP server on %s.", srv.server.Addr)

	err := srv.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Errorf("Error running localhost server: %v", err)
		Shutdown(ERROR)
		return
	}
}

type functionRequest struct {
	actionId   string
	instanceId string
	namespace  string
	step       int
	deadline   time.Time
	input      []byte
	files      []*functionFiles
	iterator   int
	fnCtx      enginerefactor.FunctionContext
}

type functionFiles struct {
	Key         string `json:"key"`
	As          string `json:"as"`
	Scope       string `json:"scope"`
	Type        string `json:"type"`
	Permissions string `json:"permissions"`
}

const sharedDir = "/mnt/shared"

type varClient interface {
	CloseSend() error
}

type varClientMsg interface {
	GetTotalSize() int64
	GetData() []byte
}

func (srv *LocalServer) requestVar(ctx context.Context, ir *functionRequest, scope, key string) (client varClient, recv func() (varClientMsg, error), err error) {
	switch scope {
	case util.VarScopeFileSystem:
		var nvClient grpc.Internal_NamespaceVariableParcelsClient
		nvClient, err = srv.flow.FileVariableParcels(ctx, &grpc.VariableInternalRequest{
			Instance: ir.instanceId,
			Key:      key,
		})
		client = nvClient
		recv = func() (varClientMsg, error) {
			return nvClient.Recv()
		}
	case util.VarScopeNamespace:
		var nvClient grpc.Internal_NamespaceVariableParcelsClient
		nvClient, err = srv.flow.NamespaceVariableParcels(ctx, &grpc.VariableInternalRequest{
			Instance: ir.instanceId,
			Key:      key,
		})
		client = nvClient
		recv = func() (varClientMsg, error) {
			return nvClient.Recv()
		}

	case util.VarScopeWorkflow:
		var wvClient grpc.Internal_WorkflowVariableParcelsClient
		wvClient, err = srv.flow.WorkflowVariableParcels(ctx, &grpc.VariableInternalRequest{
			Instance: ir.instanceId,
			Key:      key,
		})
		client = wvClient
		recv = func() (varClientMsg, error) {
			return wvClient.Recv()
		}

	case "":
		fallthrough

	case util.VarScopeInstance:
		var ivClient grpc.Internal_InstanceVariableParcelsClient
		ivClient, err = srv.flow.InstanceVariableParcels(ctx, &grpc.VariableInternalRequest{
			Instance: ir.instanceId,
			Key:      key,
		})
		client = ivClient
		recv = func() (varClientMsg, error) {
			return ivClient.Recv()
		}

	default:
		panic(scope)
	}

	return
}

type varSetClient interface {
	CloseAndRecv() (*grpc.SetVariableInternalResponse, error)
}

type varSetClientMsg struct {
	Key       string
	Instance  string
	Value     []byte
	TotalSize int64
}

func (srv *LocalServer) setVar(ctx context.Context, ir *functionRequest, totalSize int64, r io.Reader, scope, key, vMimeType string) error {
	var err error
	var client varSetClient
	var send func(*varSetClientMsg) error

	switch scope {
	case util.VarScopeFileSystem:
		return errors.New("file-system variables are read-only")
	case util.VarScopeNamespace:
		var nvClient grpc.Internal_SetNamespaceVariableParcelsClient
		nvClient, err = srv.flow.SetNamespaceVariableParcels(ctx)
		if err != nil {
			return err
		}

		client = nvClient
		send = func(x *varSetClientMsg) error {
			req := &grpc.SetVariableInternalRequest{}
			req.Key = x.Key
			req.Instance = x.Instance
			req.TotalSize = x.TotalSize
			req.Data = x.Value
			req.MimeType = vMimeType
			return nvClient.Send(req)
		}

	case util.VarScopeWorkflow:
		var wvClient grpc.Internal_SetWorkflowVariableParcelsClient
		wvClient, err = srv.flow.SetWorkflowVariableParcels(ctx)
		if err != nil {
			return err
		}

		client = wvClient
		send = func(x *varSetClientMsg) error {
			req := &grpc.SetVariableInternalRequest{}
			req.Key = x.Key
			req.Instance = x.Instance
			req.TotalSize = x.TotalSize
			req.Data = x.Value
			req.MimeType = vMimeType
			return wvClient.Send(req)
		}

	case "":
		fallthrough

	case util.VarScopeInstance:
		var ivClient grpc.Internal_SetInstanceVariableParcelsClient
		ivClient, err = srv.flow.SetInstanceVariableParcels(ctx)
		if err != nil {
			return err
		}

		client = ivClient
		send = func(x *varSetClientMsg) error {
			req := &grpc.SetVariableInternalRequest{}
			req.Key = x.Key
			req.Instance = x.Instance
			req.TotalSize = x.TotalSize
			req.Data = x.Value
			req.MimeType = vMimeType
			return ivClient.Send(req)
		}

	default:
		panic(scope)
	}

	chunkSize := int64(0x200000) // 2 MiB
	if totalSize <= 0 {
		buf := new(bytes.Buffer)
		_, err := io.CopyN(buf, r, chunkSize+1)
		if err == nil {
			return errors.New("large payload requires defined Content-Length")
		}
		if !errors.Is(err, io.EOF) {
			return err
		}

		data := buf.Bytes()
		r = bytes.NewReader(data)
		totalSize = int64(len(data))
	}

	var written int64
	for {
		chunk := chunkSize
		if totalSize-written < chunk {
			chunk = totalSize - written
		}

		buf := new(bytes.Buffer)
		k, err := io.CopyN(buf, r, chunk)
		if err != nil {
			return err
		}

		written += k

		err = send(&varSetClientMsg{
			TotalSize: totalSize,
			Key:       key,
			Instance:  ir.instanceId,
			Value:     buf.Bytes(),
		})
		if err != nil {
			return err
		}

		if written == totalSize {
			break
		}
	}

	_, err = client.CloseAndRecv()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func (srv *LocalServer) getVar(ctx context.Context, ir *functionRequest, w io.Writer, setTotalSize func(x int64), scope, key string) error {
	client, recv, err := srv.requestVar(ctx, ir, scope, key)
	if err != nil {
		return err
	}

	var received int64
	noEOF := true
	for noEOF {
		msg, err := recv()
		if errors.Is(err, io.EOF) {
			noEOF = false
		} else if err != nil {
			return err
		}

		if msg == nil {
			continue
		}

		totalSize := msg.GetTotalSize()

		if setTotalSize != nil {
			setTotalSize(totalSize)
			setTotalSize = nil
		}

		data := msg.GetData()
		received += int64(len(data))

		if received > totalSize {
			return errors.New("variable returned too many bytes")
		}

		_, err = io.Copy(w, bytes.NewReader(data))
		if err != nil {
			return err
		}

		if totalSize == received {
			break
		}
	}

	err = client.CloseSend()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
