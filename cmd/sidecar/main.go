package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/resolver"

	"github.com/vorteil/direktiv/pkg/direktiv"
	dlog "github.com/vorteil/direktiv/pkg/dlog"
	dblog "github.com/vorteil/direktiv/pkg/dlog/db"
	"github.com/vorteil/direktiv/pkg/flow"
)

const (
	exKey   = "/var/secret/exchangeKey"
	db      = "/var/secret/db"
	svcAddr = "http://localhost:8080"
)

type responseInfo struct {
	iid, aid string
	ec, em   string
	step     int32
	timeout  int
	data     []byte
	logger   dlog.Logger
}

type direktivHTTPRequest struct {
	logger    dlog.Logger
	ctxCancel context.CancelFunc
	info      *responseInfo
}

type direktivHTTPHandler struct {
	key      string
	pingAddr string

	mtx      sync.Mutex
	mtxSetup sync.Mutex

	requests map[string]*direktivHTTPRequest

	dbLog      *dblog.Logger
	flowClient flow.DirektivFlowClient
}

func main() {

	d := &direktivHTTPHandler{
		requests: make(map[string]*direktivHTTPRequest),
	}

	if os.Getenv(direktiv.DirektivDebug) == "true" {
		log.SetLevel(logrus.DebugLevel)
	}

	k, err := ioutil.ReadFile(exKey)
	if err != nil {
		log.Errorf("can not read exchange key: %v", err)
	}

	// store the key
	d.key = string(k)

	r := router.New()
	r.POST("/", d.Base)

	// ping to keep long living requests alive
	r.GET("/ping", d.Ping)

	// logs can be post or get
	r.POST("/log", d.postLog)
	r.GET("/log", d.postLog)

	// persistent variables
	r.POST("/namespace/{key}", d.postNamespaceData)
	r.GET("/namespace/{key}", d.getNamespaceData)

	r.POST("/workflow/{key}", d.postWorkflowData)
	r.GET("/workflow/{key}", d.getWorkflowData)

	r.POST("/instance/{key}", d.postInstanceData)
	r.GET("/instance/{key}", d.getInstanceData)

	// prepare ping mechanism
	go d.pingMe()

	d.dbLog, err = setupLogging()
	if err != nil {
		log.Errorf("can not setup logging: %v", err)
	}

	s := &fasthttp.Server{
		Handler: r.Handler,
	}

	// listen for sigterm
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	// subscribe to direktiv pub/sub
	err = direktiv.SyncSubscribeTo(os.Getenv(direktiv.DBConn),
		direktiv.CancelIsolate, d.handleSub)
	if err != nil {
		log.Errorf("can not setup pub/sub: %v", err)
	}

	go func(s *fasthttp.Server) {
		<-sigs
		log.Debugf("shutting down")
		if d.dbLog != nil {
			d.dbLog.CloseConnection()
		}
		s.Shutdown()
		log.Debugf("shutting down completed")
	}(s)

	log.Infof("starting direktiv sidecar container")
	err = s.ListenAndServe(":8889")
	if err != nil {
		log.Errorf("error running server: %v", err)
	}

}

func (d *direktivHTTPHandler) handleSub(in interface{}) {

	aid, ok := in.(string)
	if !ok {
		log.Errorf("cancel data %v not valid", in)
		return
	}
	log.Infof("cancelling isolate %v", aid)

	d.mtx.Lock()
	defer d.mtx.Unlock()
	if d.requests[aid] != nil && d.requests[aid].ctxCancel != nil {
		log.Infof("calling cancelling fn for %v", aid)
		d.requests[aid].ctxCancel()
		d.requests[aid].ctxCancel = nil
	}

}

func setupLogging() (*dblog.Logger, error) {

	conn, err := ioutil.ReadFile(db)
	if err != nil {
		return nil, err
	}

	return dblog.NewLogger(string(conn))

}

func (d *direktivHTTPHandler) postLog(ctx *fasthttp.RequestCtx) {

	aid := ctx.QueryArgs().Peek("aid")

	// check if this requests actually exists
	if _, ok := d.requests[string(aid)]; !ok {
		log.Errorf("request action id does not exist")
		ctx.Response.SetStatusCode(500)
		return
	}

	// get and post is supported
	var l []byte
	if string(ctx.Method()) == "POST" {
		l = ctx.Request.Body()
	} else {
		l = ctx.QueryArgs().Peek("log")
	}

	d.mtx.Lock()
	defer d.mtx.Unlock()
	if d.requests[string(aid)].logger != nil {
		d.requests[string(aid)].logger.Info(string(l))
	}

}

func (d *direktivHTTPHandler) Ping(ctx *fasthttp.RequestCtx) {
	log.Debugf("direktiv sidecar alive ping")
	ctx.WriteString("pong")
}

func generateError(ctx *fasthttp.RequestCtx, errCode, errMsg string) {
	ctx.Response.SetStatusCode(500)
	jsonResponse := direktiv.ServiceResponse{
		ErrorCode:    errCode,
		ErrorMessage: errMsg,
	}
	b, _ := json.Marshal(jsonResponse)
	fmt.Fprintf(ctx, string(b))
}

func checkHeader(ctx *fasthttp.RequestCtx, hdr string) (string, error) {
	val := ctx.Request.Header.Peek(hdr)
	if len(val) == 0 {
		return "", fmt.Errorf("header missing: %s", hdr)
	}
	return string(val), nil
}

// Base is the main function receiving requests and handling pings/logs and
// response if required
func (d *direktivHTTPHandler) Base(ctx *fasthttp.RequestCtx) {

	// headers to check for
	hdrs := []string{direktiv.DirektivExchangeKeyHeader,
		direktiv.DirektivActionIDHeader,
		direktiv.DirektivPingAddrHeader,
		direktiv.DirektivInstanceIDHeader,
		direktiv.DirektivTimeoutHeader,
		direktiv.DirektivStepHeader,
		direktiv.DirektivResponseHeader,
		direktiv.DirektivNamespaceHeader,
		direktiv.DirektivSourceHeader,
	}

	// map with values of the headers
	vals := make(map[string]string)

	for _, j := range hdrs {
		v, err := checkHeader(ctx, j)
		if err != nil {
			generateError(ctx, direktiv.ServiceErrorInternal, err.Error())
			return
		}
		vals[j] = v
	}

	to, err := strconv.Atoi(vals[direktiv.DirektivTimeoutHeader])
	if err != nil {
		generateError(ctx, direktiv.ServiceErrorInternal,
			fmt.Sprintf("timeout form incorrect: %s", err.Error()))
		return
	}

	// reset timeout to 900 secs if 0
	if to == 0 {
		to = 900
	}

	// check that key and provided key are the same
	if d.key != vals[direktiv.DirektivExchangeKeyHeader] {
		generateError(ctx, direktiv.ServiceErrorInternal,
			fmt.Sprintf("header incorrect: %s", direktiv.DirektivExchangeKeyHeader))
		return
	}

	// step needs to be in right format
	step, err := strconv.ParseInt(vals[direktiv.DirektivStepHeader], 10, 64)
	if err != nil {
		generateError(ctx, direktiv.ServiceErrorInternal,
			fmt.Sprintf("step form incorrect: %s", err.Error()))
		return
	}

	// disable/enable ping
	d.mtxSetup.Lock()
	if len(d.pingAddr) == 0 {
		d.pingAddr = vals[direktiv.DirektivPingAddrHeader]
	}

	if d.flowClient == nil {

		conn, err := direktiv.GetEndpointTLS(vals[direktiv.DirektivResponseHeader], true)
		if err != nil {
			log.Errorf("can not connect to direktiv ingress: %v", err)
			generateError(ctx, direktiv.ServiceErrorInternal,
				fmt.Sprintf("can not setup flow client: %s", err.Error()))
			return
		}

		log.Infof("connecting to %s", vals[direktiv.DirektivResponseHeader])

		d.flowClient = flow.NewDirektivFlowClient(conn)
	}

	log15log, err := d.dbLog.LoggerFunc(vals[direktiv.DirektivNamespaceHeader],
		vals[direktiv.DirektivInstanceIDHeader])
	if err != nil {
		generateError(ctx, direktiv.ServiceErrorInternal,
			fmt.Sprintf("can not setup logger: %s", err.Error()))
		return
	}
	d.mtxSetup.Unlock()

	fmt.Fprintf(ctx, "ok")

	info := &responseInfo{
		iid:     vals[direktiv.DirektivInstanceIDHeader],
		aid:     vals[direktiv.DirektivActionIDHeader],
		step:    int32(step),
		logger:  log15log,
		data:    ctx.Request.Body(),
		timeout: to,
	}

	go d.handleSubRequest(info)

}

func (d *direktivHTTPHandler) handleSubRequest(info *responseInfo) {

	// timeout in context & client
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(info.timeout*2)*time.Second)
	defer cancel()

	// we are adding some time to the "technical" timeout
	client := &http.Client{
		Timeout: time.Duration(info.timeout+10) * time.Second,
	}

	// add to request manager
	d.mtx.Lock()
	d.requests[info.aid] = &direktivHTTPRequest{
		logger:    info.logger,
		ctxCancel: cancel,
		info:      info,
	}
	d.mtx.Unlock()

	defer func() {
		log.Debugf("cleanup request map")
		d.mtx.Lock()
		if _, ok := d.requests[info.aid]; ok {
			delete(d.requests, info.aid)
		}
		d.mtx.Unlock()
	}()

	log.Debugf("handle request aid: %s", info.aid)

	// wipe data field for "real" response
	body := info.data
	info.data = []byte{}

	// forward request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		svcAddr, bytes.NewReader(body))
	if err != nil {
		log.Errorf("can not request service: %v", err)
		info.ec = direktiv.ServiceErrorNetwork
		info.em = fmt.Sprintf("create request failed: %s", err.Error())
		d.respondToFlow(info)
		return
	}

	// add header so client can use it ass reference
	req.Header.Add(direktiv.DirektivActionIDHeader, info.aid)

	resp, err := client.Do(req)
	if err != nil {
		// only respond if it has not been cancelled
		d.mtx.Lock()
		if d.requests[info.aid] != nil &&
			d.requests[info.aid].ctxCancel != nil {
			info.ec = direktiv.ServiceErrorNetwork
			info.em = fmt.Sprintf("execute request failed: %s", err.Error())
			d.respondToFlow(info)
		}
		d.mtx.Unlock()
		return
	}

	// check if service reports an error
	if resp.Header.Get(direktiv.DirektivErrorCodeHeader) != "" {
		info.ec = resp.Header.Get(direktiv.DirektivErrorCodeHeader)
		info.em = resp.Header.Get(direktiv.DirektivErrorMessageHeader)
		d.respondToFlow(info)
		return
	}

	defer resp.Body.Close()
	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		info.ec = direktiv.ServiceErrorNetwork
		info.em = fmt.Sprintf("reading body failed: %s", err.Error())
		d.respondToFlow(info)
		return
	}

	info.data = f

	// respond to service
	d.respondToFlow(info)

}

// this keeps the knative function alive if we have long running actions
// as long as we are waiting for a request we keep ping every 10 seconds
func (d *direktivHTTPHandler) pingMe() {

	for range time.Tick(5 * time.Second) {
		if len(d.pingAddr) > 0 && d.pingAddr != "noping" && len(d.requests) > 0 {
			_, err := http.Get(fmt.Sprintf("%s/ping", d.pingAddr))
			log.Debugf("ping %s: %v", fmt.Sprintf("%s/ping", d.pingAddr), err)
		}
	}

}

func (d *direktivHTTPHandler) respondToFlow(info *responseInfo) {

	r := &flow.ReportActionResultsRequest{
		InstanceId:   &info.iid,
		Step:         &info.step,
		ActionId:     &info.aid,
		Output:       info.data,
		ErrorCode:    &info.ec,
		ErrorMessage: &info.em,
	}

	_, err := d.flowClient.ReportActionResults(context.Background(), r)
	if err != nil {
		log.Errorf("can not respond to flow: %v", err)
		return
	}

}

func init() {
	resolver.Register(&direktiv.KubeResolverBuilder{})
}

const grpcChunkSize = 2 * 1024 * 1024

func (d *direktivHTTPHandler) postNamespaceData(ctx *fasthttp.RequestCtx) {

	aid := ctx.QueryArgs().Peek("aid")

	// check if this requests actually exists
	var info *responseInfo
	d.mtx.Lock()
	if r, ok := d.requests[string(aid)]; ok {
		info = r.info
	} else {
		log.Errorf("request action id does not exist")
		ctx.Response.SetStatusCode(500)
		return
	}
	d.mtx.Unlock()

	key := ctx.UserValue("key").(string)

	str, err := checkHeader(ctx, "Content-Length")
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	l, err := strconv.Atoi(str)
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	totalSize := int64(l)
	chunkSize := int64(grpcChunkSize)

	r := ctx.RequestBodyStream()

	client, err := d.flowClient.SetNamespaceVariable(context.Background())
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	var totalRead int64
	var chunks int

	for {
		buf := new(bytes.Buffer)
		rdr := io.LimitReader(r, chunkSize)
		var k int64
		k, err = io.Copy(buf, rdr)
		totalRead += k
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}

		if k == 0 && chunks > 0 {
			break
		}

		data := buf.Bytes()

		req := new(flow.SetNamespaceVariableRequest)
		req.InstanceId = &info.iid
		req.Key = &key
		req.Value = data
		req.TotalSize = &totalSize
		req.ChunkSize = &chunkSize

		err = client.Send(req)
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}

		chunks++
		if totalRead >= totalSize {
			break
		}
	}

}

func (d *direktivHTTPHandler) postWorkflowData(ctx *fasthttp.RequestCtx) {

	aid := ctx.QueryArgs().Peek("aid")

	// check if this requests actually exists
	var info *responseInfo
	d.mtx.Lock()
	if r, ok := d.requests[string(aid)]; ok {
		info = r.info
	} else {
		log.Errorf("request action id does not exist")
		ctx.Response.SetStatusCode(500)
		return
	}
	d.mtx.Unlock()

	key := ctx.UserValue("key").(string)

	str, err := checkHeader(ctx, "Content-Length")
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	l, err := strconv.Atoi(str)
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	totalSize := int64(l)
	chunkSize := int64(grpcChunkSize)

	r := ctx.RequestBodyStream()

	client, err := d.flowClient.SetWorkflowVariable(context.Background())
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	var totalRead int64
	var chunks int

	for {
		buf := new(bytes.Buffer)
		rdr := io.LimitReader(r, chunkSize)
		var k int64
		k, err = io.Copy(buf, rdr)
		totalRead += k
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}

		if k == 0 && chunks > 0 {
			break
		}

		data := buf.Bytes()

		req := new(flow.SetWorkflowVariableRequest)
		req.InstanceId = &info.iid
		req.Key = &key
		req.Value = data
		req.TotalSize = &totalSize
		req.ChunkSize = &chunkSize

		err = client.Send(req)
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}

		chunks++
		if totalRead >= totalSize {
			break
		}
	}

}

func (d *direktivHTTPHandler) postInstanceData(ctx *fasthttp.RequestCtx) {

	aid := ctx.QueryArgs().Peek("aid")

	// check if this requests actually exists
	var info *responseInfo
	d.mtx.Lock()
	if r, ok := d.requests[string(aid)]; ok {
		info = r.info
	} else {
		log.Errorf("request action id does not exist")
		ctx.Response.SetStatusCode(500)
		return
	}
	d.mtx.Unlock()

	key := ctx.UserValue("key").(string)

	str, err := checkHeader(ctx, "Content-Length")
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	l, err := strconv.Atoi(str)
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	totalSize := int64(l)
	chunkSize := int64(grpcChunkSize)

	r := ctx.RequestBodyStream()

	client, err := d.flowClient.SetInstanceVariable(context.Background())
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	var totalRead int64
	var chunks int

	for {
		buf := new(bytes.Buffer)
		rdr := io.LimitReader(r, chunkSize)
		var k int64
		k, err = io.Copy(buf, rdr)
		totalRead += k
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}

		if k == 0 && chunks > 0 {
			break
		}

		data := buf.Bytes()

		req := new(flow.SetInstanceVariableRequest)
		req.InstanceId = &info.iid
		req.Key = &key
		req.Value = data
		req.TotalSize = &totalSize
		req.ChunkSize = &chunkSize

		err = client.Send(req)
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}

		chunks++
		if totalRead >= totalSize {
			break
		}
	}

}

func (d *direktivHTTPHandler) getNamespaceData(ctx *fasthttp.RequestCtx) {

	aid := ctx.QueryArgs().Peek("aid")

	// check if this requests actually exists
	var info *responseInfo
	d.mtx.Lock()
	if r, ok := d.requests[string(aid)]; ok {
		info = r.info
	} else {
		log.Errorf("request action id does not exist")
		ctx.Response.SetStatusCode(500)
		return
	}
	d.mtx.Unlock()

	key := ctx.UserValue("key").(string)

	req := &flow.GetNamespaceVariableRequest{
		InstanceId: &info.iid,
		Key:        &key,
	}

	client, err := d.flowClient.GetNamespaceVariable(context.Background(), req)
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	in, err := client.Recv()
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	// chunkSize := in.GetChunkSize()
	// totalSize := in.GetTotalSize()

	for {
		k, err := io.Copy(ctx, bytes.NewReader(in.GetValue()))
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}
		if k == 0 {
			break
		}
	}

}

func (d *direktivHTTPHandler) getWorkflowData(ctx *fasthttp.RequestCtx) {

	aid := ctx.QueryArgs().Peek("aid")

	// check if this requests actually exists
	var info *responseInfo
	d.mtx.Lock()
	if r, ok := d.requests[string(aid)]; ok {
		info = r.info
	} else {
		log.Errorf("request action id does not exist")
		ctx.Response.SetStatusCode(500)
		return
	}
	d.mtx.Unlock()

	key := ctx.UserValue("key").(string)

	req := &flow.GetWorkflowVariableRequest{
		InstanceId: &info.iid,
		Key:        &key,
	}

	client, err := d.flowClient.GetWorkflowVariable(context.Background(), req)
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	in, err := client.Recv()
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	// chunkSize := in.GetChunkSize()
	// totalSize := in.GetTotalSize()

	for {
		k, err := io.Copy(ctx, bytes.NewReader(in.GetValue()))
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}
		if k == 0 {
			break
		}
	}

}

func (d *direktivHTTPHandler) getInstanceData(ctx *fasthttp.RequestCtx) {

	aid := ctx.QueryArgs().Peek("aid")

	// check if this requests actually exists
	var info *responseInfo
	d.mtx.Lock()
	if r, ok := d.requests[string(aid)]; ok {
		info = r.info
	} else {
		log.Errorf("request action id does not exist")
		ctx.Response.SetStatusCode(500)
		return
	}
	d.mtx.Unlock()

	key := ctx.UserValue("key").(string)

	req := &flow.GetInstanceVariableRequest{
		InstanceId: &info.iid,
		Key:        &key,
	}

	client, err := d.flowClient.GetInstanceVariable(context.Background(), req)
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	in, err := client.Recv()
	if err != nil {
		log.Error(err)
		ctx.Response.SetStatusCode(500)
		return
	}

	// chunkSize := in.GetChunkSize()
	// totalSize := in.GetTotalSize()

	for {
		k, err := io.Copy(ctx, bytes.NewReader(in.GetValue()))
		if err != nil {
			log.Error(err)
			ctx.Response.SetStatusCode(500)
			return
		}
		if k == 0 {
			break
		}
	}

}
