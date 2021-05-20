package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/flow"
	"google.golang.org/protobuf/types/known/emptypb"

	log "github.com/sirupsen/logrus"
	dblog "github.com/vorteil/direktiv/pkg/dlog/db"
)

const (
	workerThreads = 10
)

type LocalServer struct {
	end     func()
	flow    flow.DirektivFlowClient
	logging *dblog.Logger
	queue   chan *inboundRequest
	router  *mux.Router
	stopper chan *time.Time
	server  http.Server
	workers []*inboundWorker

	requestsLock sync.Mutex
	requests     map[string]*activeRequest
}

func (srv *LocalServer) initFlow() error {

	flowAddr := os.Getenv("DIREKTIV_FLOW_ENDPOINT")

	log.Infof("Connecting to flow: %s.", flowAddr)

	conn, err := direktiv.GetEndpointTLS(flowAddr, true)
	if err != nil {
		return err
	}

	srv.flow = flow.NewDirektivFlowClient(conn)

	return nil

}

func (srv *LocalServer) initLogging() error {

	var err error

	conn := os.Getenv("DIREKTIV_DB")

	log.Infof("Connecting to instance logs database.")
	log.Debugf("Instance logs database connection string: %s", string(conn))

	srv.logging, err = dblog.NewLogger(string(conn))
	if err != nil {
		return err
	}

	return nil

}

func (srv *LocalServer) initPubSub() error {

	addr := os.Getenv("DIREKTIV_DB")

	log.Infof("Connecting to pub/sub service.")
	log.Debugf("Pub/sub connection string: %s", addr)

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

	err = srv.initLogging()
	if err != nil {
		log.Errorf("Localhost server unable to connect to instance logging: %v", err)
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

	req.logger.Info(msg)

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

type inboundWorker struct {
	id     int
	cancel func()
	lock   sync.Mutex
	srv    *LocalServer
}

func (worker *inboundWorker) Cancel() {

	worker.lock.Lock()

	if worker.cancel != nil {
		log.Debugf("Cancelling worker %d.", worker.id)
		worker.cancel()
	}

	worker.lock.Unlock()

}

func (worker *inboundWorker) run() {

	log.Debugf("Starting worker %d.", worker.id)

	for {
		worker.lock.Lock()

		req, more := <-worker.srv.queue
		if req == nil || !more {
			worker.cancel = nil
			worker.lock.Unlock()
			break
		}

		ctx, cancel := context.WithCancel(req.r.Context())
		worker.cancel = cancel
		req.r = req.r.WithContext(ctx)

		worker.lock.Unlock()

		id := req.r.Header.Get(actionIDHeader)
		log.Debugf("Worker %d picked up request '%s'.", worker.id, id)

		worker.handleIsolateRequest(req)

	}

	log.Debugf("Worker %d shut down.", worker.id)

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
	logger     dlog.Logger
}

type isolateFiles struct {
	Key   string `json:"key"`
	As    string `json:"as"`
	Scope string `json:"scope"`
	Type  string `json:"type"`
}

func (worker *inboundWorker) handleIsolateRequest(req *inboundRequest) {

	defer func() {
		close(req.end)
	}()

	ir := worker.validateIsolateRequest(req)
	if ir == nil {
		return
	}
	defer ir.logger.Close()

	ctx := req.r.Context()
	ctx, cancel := context.WithDeadline(ctx, ir.deadline)
	defer cancel()

	reportError := func(err error) {

		log.Warnf("Action '%s' returning uncatchable error: %v.", ir.actionId, err)

		ctx := context.Background() // TODO

		worker.respondToFlow(ctx, ir, &outcome{
			errMsg: err.Error(),
		})

	}

	defer worker.cleanupIsolateRequest(ir)

	err := worker.prepIsolateRequest(ctx, ir)
	if err != nil {
		reportError(err)
		return
	}

	// NOTE: rctx exists because we don't want to immediately cancel the isolate request if our context is cancelled
	rctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.srv.registerActiveRequest(ir, rctx, cancel)
	defer worker.srv.deregisterActiveRequest(ir.actionId)
	go func() {
		select {
		case <-rctx.Done():
		case <-ctx.Done():
			worker.srv.cancelActiveRequest(rctx, ir.actionId)
		}
	}()

	out, err := worker.doIsolateRequest(rctx, ir)
	if err != nil {
		reportError(err)
		return
	}

	worker.respondToFlow(rctx, ir, out)

}

func (worker *inboundWorker) respondToFlow(ctx context.Context, ir *isolateRequest, out *outcome) {

	step := int32(ir.step)

	_, err := worker.srv.flow.ReportActionResults(ctx, &flow.ReportActionResultsRequest{
		InstanceId:   &ir.instanceId,
		Step:         &step,
		ActionId:     &ir.actionId,
		Output:       out.data,
		ErrorCode:    &out.errCode,
		ErrorMessage: &out.errMsg,
	})

	if err != nil {
		log.Errorf("Failed to report results for request '%s': %v.", ir.actionId, err)
		return
	}

	if out.errCode != "" {
		log.Infof("Request '%s' failed with catchable error '%s': %s.", ir.actionId, out.errCode, out.errMsg)
	} else if out.errMsg != "" {
		log.Infof("Request '%s' failed with uncatchable service error: %s.", ir.actionId, out.errMsg)
	} else {
		log.Infof("Request '%s' completed successfully.", ir.actionId)
	}

}

func (worker *inboundWorker) validateIsolateRequest(req *inboundRequest) *isolateRequest {

	id := req.r.Header.Get(actionIDHeader)

	reportError := func(code int, err error) {

		msg := err.Error()

		http.Error(req.w, msg, code)

		log.Warnf("Request '%s' returned %v due to failed validation: %v.", id, code, err)

	}

	var err error

	ir := new(isolateRequest)

	hdr := actionIDHeader
	s := req.r.Header.Get(hdr)
	ir.actionId = s
	if s == "" {
		reportError(http.StatusBadRequest, fmt.Errorf("missing %s", hdr))
		return nil
	}

	hdr = "Direktiv-InstanceID"
	s = req.r.Header.Get(hdr)
	ir.instanceId = s
	if s == "" {
		reportError(http.StatusBadRequest, fmt.Errorf("missing %s", hdr))
		return nil
	}

	hdr = "Direktiv-Namespace"
	s = req.r.Header.Get(hdr)
	ir.namespace = s
	if s == "" {
		reportError(http.StatusBadRequest, fmt.Errorf("missing %s", hdr))
		return nil
	}

	hdr = "Direktiv-Step"
	s = req.r.Header.Get(hdr)
	ir.step, err = strconv.Atoi(s)
	if err != nil {
		reportError(http.StatusBadRequest, fmt.Errorf("invalid %s: %v", hdr, err))
		return nil
	}
	if ir.step < 0 {
		reportError(http.StatusBadRequest, fmt.Errorf("invalid %s value: %v", hdr, s))
		return nil
	}

	hdr = "Direktiv-Deadline"
	s = req.r.Header.Get(hdr)
	ir.deadline, err = time.Parse(time.RFC3339, s)
	if err != nil {
		reportError(http.StatusBadRequest, fmt.Errorf("invalid %s: %v", hdr, err))
		return nil
	}

	cap := int64(0x400000) // 4 MiB
	if req.r.ContentLength == 0 {
		code := http.StatusLengthRequired
		reportError(code, errors.New(http.StatusText(code)))
		return nil
	}
	if req.r.ContentLength > cap {
		reportError(http.StatusRequestEntityTooLarge, fmt.Errorf("size limit: %d bytes", cap))
		return nil
	}
	r := io.LimitReader(req.r.Body, cap)

	ir.input, err = ioutil.ReadAll(r)
	if err != nil {
		reportError(http.StatusBadRequest, fmt.Errorf("failed to read request body: %v", err))
		return nil
	}
	if int64(len(ir.input)) != req.r.ContentLength {
		reportError(http.StatusBadRequest, fmt.Errorf("request body doesn't match Content-Length"))
		return nil
	}

	hdr = "Direktiv-Files"
	strs := req.r.Header.Values(hdr)
	for i, s := range strs {

		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			reportError(http.StatusBadRequest, fmt.Errorf("invalid %s [%d]: %v", hdr, i, err))
			return nil
		}

		files := new(isolateFiles)
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		err = dec.Decode(files)
		if err != nil {
			reportError(http.StatusBadRequest, fmt.Errorf("invalid %s [%d]: %v", hdr, i, err))
			return nil
		}

		// TODO: extra validation

		ir.files = append(ir.files, files)

	}

	ir.logger, err = worker.srv.logging.LoggerFunc(ir.namespace, ir.instanceId)
	if err != nil {
		code := http.StatusInternalServerError
		msg := http.StatusText(code)
		reportError(code, errors.New(msg))
		return nil
	}

	return ir

}

const sharedDir = "/var/log"

func (worker *inboundWorker) isolateDir(ir *isolateRequest) string {
	return filepath.Join(sharedDir, ir.actionId)
}

func (worker *inboundWorker) cleanupIsolateRequest(ir *isolateRequest) {
	dir := worker.isolateDir(ir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.Error(err)
	}
}

func (worker *inboundWorker) prepIsolateRequest(ctx context.Context, ir *isolateRequest) error {

	err := worker.prepIsolateFiles(ctx, ir)
	if err != nil {
		return fmt.Errorf("failed to prepare isolate files: %v", err)
	}

	return nil

}

func (worker *inboundWorker) prepIsolateFiles(ctx context.Context, ir *isolateRequest) error {

	dir := worker.isolateDir(ir)

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	for i, f := range ir.files {
		err = worker.prepOneIsolateFiles(ctx, ir, f)
		if err != nil {
			return fmt.Errorf("failed to prepare isolate files %d: %v", i, err)
		}
	}

	return nil

}

type varClient interface {
	CloseSend() error
}

type varClientMsg interface {
	GetTotalSize() int64
	GetChunkSize() int64
	GetValue() []byte
}

func (worker *inboundWorker) prepOneIsolateFiles(ctx context.Context, ir *isolateRequest, f *isolateFiles) error {

	pr, pw := io.Pipe()

	go func() {
		err := worker.fileReader(ctx, ir, f, pw)
		if err != nil {
			_ = pw.CloseWithError(err)
		} else {
			_ = pw.Close()
		}
	}()

	err := worker.fileWriter(ctx, ir, f, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return err
	}

	_ = pr.Close()

	return nil

}

func untar(dst string, r io.Reader) error {

	err := os.MkdirAll(dst, 0777)
	if err != nil {
		return err
	}

	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dst, hdr.Name)

		if hdr.Typeflag == tar.TypeReg {
			pdir, _ := filepath.Split(path)
			err = os.MkdirAll(pdir, 0777)
			if err != nil {
				return err
			}

			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, tr)
			if err != nil {
				return err
			}

			err = f.Close()
			if err != nil {
				return err
			}

		} else if hdr.Typeflag == tar.TypeDir {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unsupported tar archive contents")
		}

	}

	return nil

}

func (worker *inboundWorker) fileWriter(ctx context.Context, ir *isolateRequest, f *isolateFiles, pr *io.PipeReader) error {

	// TODO: const the types
	// TODO: validate f.Type earlier so that the switch cannot get unexpected data here

	dir := worker.isolateDir(ir)
	dst := f.Key
	if f.As != "" {
		dst = f.As
	}
	dst = filepath.Join(dir, dst)
	dir, _ = filepath.Split(dst)

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	switch f.Type {

	case "":
		fallthrough

	case "plain":
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, pr)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}

	case "base64":
		r := base64.NewDecoder(base64.StdEncoding, pr)

		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, r)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}

	case "tar":

		err = untar(dst, pr)
		if err != nil {
			return err
		}

	case "tar.gz":

		gr, err := gzip.NewReader(pr)
		if err != nil {
			return err
		}

		err = untar(dst, gr)
		if err != nil {
			return err
		}

		err = gr.Close()
		if err != nil {
			return err
		}

	default:
		panic(f.Type)
	}

	return nil

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
		ivClient.Recv()
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
	if totalSize == 0 {
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

		if totalSize > received {
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

func (worker *inboundWorker) fileReader(ctx context.Context, ir *isolateRequest, f *isolateFiles, pw *io.PipeWriter) error {

	defer pw.Close()

	err := worker.srv.getVar(ctx, ir, pw, nil, f.Scope, f.Key)
	if err != nil {
		return err
	}

	return nil

}

type outcome struct {
	data    []byte
	errCode string
	errMsg  string
}

func (worker *inboundWorker) doIsolateRequest(ctx context.Context, ir *isolateRequest) (*outcome, error) {

	log.Debugf("Forwarding request '%s' to service.", ir.actionId)

	url := "http://localhost:8080"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(ir.input))
	if err != nil {
		return nil, err
	}

	req.Header.Set(actionIDHeader, ir.actionId)
	req.Header.Set("Direktiv-TempDir", worker.isolateDir(ir))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out := new(outcome)

	out.errCode = resp.Header.Get("Direktiv-ErrorCode")
	out.errMsg = resp.Header.Get("Direktiv-ErrorMessage")

	if out.errCode != "" {
		return out, nil
	}

	cap := int64(0x400000) // 4 MiB
	if resp.ContentLength > cap {
		return nil, errors.New("service response is too large")
	}
	r := io.LimitReader(resp.Body, cap)

	out.data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return out, nil

}
