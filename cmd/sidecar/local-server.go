package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/flow"
	"github.com/vorteil/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"

	log "github.com/sirupsen/logrus"
)

const (
	workerThreads = 10
)

type LocalServer struct {
	end     func()
	flow    flow.DirektivFlowClient
	queue   chan *inboundRequest
	router  *mux.Router
	stopper chan *time.Time
	server  http.Server
	workers []*inboundWorker

	requestsLock sync.Mutex
	requests     map[string]*activeRequest
}

func (srv *LocalServer) initFlow() error {

	conn, err := util.GetEndpointTLS(util.TLSFlowComponent)
	if err != nil {
		return err
	}

	log.Debugf(">> conn %v", conn)

	srv.flow = flow.NewDirektivFlowClient(conn)

	return nil

}

func (srv *LocalServer) initPubSub() error {

	addr := os.Getenv("DIREKTIV_DB")

	log.Infof("Connecting to pub/sub service.")

	err := direktiv.SyncSubscribeTo(addr,
		direktiv.CancelIsolate, srv.handlePubSubCancel)
	if err != nil {
		return err
	}

	return nil

}

func (srv *LocalServer) handlePubSubCancel(in interface{}) {

	actionId, ok := in.(string)
	if !ok {
		log.Errorf("cancel data %v not valid", in)
		return
	}
	log.Infof("cancelling isolate %v", actionId)

	// TODO: do we need to find a better way to cancel requests that come late off the queue?

	srv.cancelActiveRequest(context.Background(), actionId)

}

func (srv *LocalServer) Start() {

	err := srv.initFlow()
	if err != nil {
		log.Errorf("Localhost server unable to connect to flow: %v", err)
		Shutdown(ERROR)
		return
	}

	err = srv.initPubSub()
	if err != nil {
		log.Errorf("Localhost server unable to set up pub/sub: %v", err)
		Shutdown(ERROR)
		return
	}

	srv.queue = make(chan *inboundRequest, 100)
	srv.requests = make(map[string]*activeRequest)

	srv.router = mux.NewRouter()
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
	req, _ := srv.requests[actionId]
	srv.requestsLock.Unlock()

	reportError := func(code int, err error) {
		http.Error(w, err.Error(), code)
		log.Warnf("Log handler for '%s' returned %v: %v.", actionId, code, err)
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

		data, err := ioutil.ReadAll(r)
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

	_, err := srv.flow.ActionLog(req.ctx, &flow.ActionLogRequest{
		InstanceId: &req.instanceId,
		Msg:        []string{msg},
	})
	if err != nil {
		log.Errorf("Failed to forward log to diretiv: %v.", err)
	}

	log.Debugf("Log handler for '%s' posted %d bytes.", actionId, len(msg))

}

func (srv *LocalServer) varHandler(w http.ResponseWriter, r *http.Request) {

	actionId := r.URL.Query().Get("aid")

	srv.requestsLock.Lock()
	req, _ := srv.requests[actionId]
	srv.requestsLock.Unlock()

	reportError := func(code int, err error) {
		http.Error(w, err.Error(), code)
		log.Warnf("Var handler for '%s' returned %v: %v.", actionId, code, err)
		return
	}

	if req == nil {
		code := http.StatusNotFound
		reportError(code, fmt.Errorf("actionId %s not found", actionId))
		return
	}

	ctx := req.ctx
	ir := req.isolateRequest

	scope := r.URL.Query().Get("scope")
	key := r.URL.Query().Get("key")

	switch r.Method {
	case http.MethodGet:

		setTotalSize := func(x int64) {
			w.Header().Set("Content-Length", strconv.Itoa(int(x)))
		}

		err := srv.getVar(ctx, ir, w, setTotalSize, scope, key)
		if err != nil {
			// TODO: more specific errors
			reportError(http.StatusInternalServerError, err)
			return
		}

		log.Debugf("Var handler for '%s' retrieved %s (%s)", actionId, key, scope)

	case http.MethodPost:

		err := srv.setVar(ctx, ir, r.ContentLength, r.Body, scope, key)
		if err != nil {
			// TODO: more specific errors
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
	*isolateRequest
	cancel func()
	ctx    context.Context
}

func (srv *LocalServer) registerActiveRequest(ir *isolateRequest, ctx context.Context, cancel func()) {

	srv.requestsLock.Lock()

	srv.requests[ir.actionId] = &activeRequest{
		isolateRequest: ir,
		ctx:            ctx,
		cancel:         cancel,
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

	req, _ := srv.requests[actionId]

	srv.requestsLock.Unlock()

	if req == nil {
		return
	}

	log.Infof("Attempting to cancel '%s'.", actionId)

	go srv.sendCancelToService(ctx, req.isolateRequest)

	select {
	case <-req.ctx.Done():
	case <-time.After(10 * time.Second):
		log.Warnf("Request '%s' failed to cancel punctually.", actionId)
		req.cancel()
	}

}

func (srv *LocalServer) sendCancelToService(ctx context.Context, ir *isolateRequest) {

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

type isolateRequest struct {
	actionId   string
	instanceId string
	namespace  string
	step       int
	deadline   time.Time
	input      []byte
	files      []*isolateFiles
	errCode    string
	errMsg     string
}

type isolateFiles struct {
	Key   string `json:"key"`
	As    string `json:"as"`
	Scope string `json:"scope"`
	Type  string `json:"type"`
}

const sharedDir = "/mnt/shared"

type varClient interface {
	CloseSend() error
}

type varClientMsg interface {
	GetTotalSize() int64
	GetChunkSize() int64
	GetValue() []byte
}

func (srv *LocalServer) requestVar(ctx context.Context, ir *isolateRequest, scope, key string) (client varClient, recv func() (varClientMsg, error), err error) {

	// TODO: const the scopes
	// TODO: validate scope earlier so that the switch cannot get unexpected data here
	// TODO: log missing files but proceed anyway

	switch scope {

	case "namespace":
		var nvClient flow.DirektivFlow_GetNamespaceVariableClient
		nvClient, err = srv.flow.GetNamespaceVariable(ctx, &flow.GetNamespaceVariableRequest{
			InstanceId: &ir.instanceId,
			Key:        &key,
		})
		client = nvClient
		recv = func() (varClientMsg, error) {
			return nvClient.Recv()
		}

	case "workflow":
		var wvClient flow.DirektivFlow_GetWorkflowVariableClient
		wvClient, err = srv.flow.GetWorkflowVariable(ctx, &flow.GetWorkflowVariableRequest{
			InstanceId: &ir.instanceId,
			Key:        &key,
		})
		client = wvClient
		recv = func() (varClientMsg, error) {
			return wvClient.Recv()
		}

	case "":
		fallthrough

	case "instance":
		var ivClient flow.DirektivFlow_GetInstanceVariableClient
		ivClient, err = srv.flow.GetInstanceVariable(ctx, &flow.GetInstanceVariableRequest{
			InstanceId: &ir.instanceId,
			Key:        &key,
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
	CloseAndRecv() (*emptypb.Empty, error)
}

type varSetClientMsg struct {
	Key        *string
	InstanceId *string
	Value      []byte
	TotalSize  *int64
	ChunkSize  *int64
}

func (srv *LocalServer) setVar(ctx context.Context, ir *isolateRequest, totalSize int64, r io.Reader, scope, key string) error {

	// TODO: const the scopes
	// TODO: validate scope earlier so that the switch cannot get unexpected data here
	// TODO: log missing files but proceed anyway

	var err error
	var client varSetClient
	var send func(*varSetClientMsg) error

	switch scope {

	case "namespace":
		var nvClient flow.DirektivFlow_SetNamespaceVariableClient
		nvClient, err = srv.flow.SetNamespaceVariable(ctx)
		client = nvClient
		send = func(x *varSetClientMsg) error {
			req := &flow.SetNamespaceVariableRequest{}
			req.Key = x.Key
			req.InstanceId = x.InstanceId
			req.TotalSize = x.TotalSize
			req.Value = x.Value
			req.ChunkSize = x.ChunkSize
			return nvClient.Send(req)
		}

	case "workflow":
		var wvClient flow.DirektivFlow_SetWorkflowVariableClient
		wvClient, err = srv.flow.SetWorkflowVariable(ctx)
		client = wvClient
		send = func(x *varSetClientMsg) error {
			req := &flow.SetWorkflowVariableRequest{}
			req.Key = x.Key
			req.InstanceId = x.InstanceId
			req.TotalSize = x.TotalSize
			req.Value = x.Value
			req.ChunkSize = x.ChunkSize
			return wvClient.Send(req)
		}

	case "":
		fallthrough

	case "instance":
		var ivClient flow.DirektivFlow_SetInstanceVariableClient
		ivClient, err = srv.flow.SetInstanceVariable(ctx)
		client = ivClient
		send = func(x *varSetClientMsg) error {
			req := &flow.SetInstanceVariableRequest{}
			req.Key = x.Key
			req.InstanceId = x.InstanceId
			req.TotalSize = x.TotalSize
			req.Value = x.Value
			req.ChunkSize = x.ChunkSize
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
		if err != io.EOF {
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
			TotalSize:  &totalSize,
			ChunkSize:  &chunkSize,
			Key:        &key,
			InstanceId: &ir.instanceId,
			Value:      buf.Bytes(),
		})
		if err != nil {
			return err
		}

		if written == totalSize {
			break
		}
	}

	_, err = client.CloseAndRecv()
	if err != nil && err != io.EOF {
		return err
	}

	return nil

}

func (srv *LocalServer) getVar(ctx context.Context, ir *isolateRequest, w io.Writer, setTotalSize func(x int64), scope, key string) error {

	client, recv, err := srv.requestVar(ctx, ir, scope, key)
	if err != nil {
		return err
	}

	var received int64
	var noEOF = true
	for noEOF {
		msg, err := recv()
		if err == io.EOF {
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

		data := msg.GetValue()
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
	if err != nil && err != io.EOF {
		return err
	}

	return nil

}
