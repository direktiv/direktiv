package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"google.golang.org/grpc"

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

type direktivHTTPRequest struct {
	ctx    context.Context
	logger dlog.Logger
}

type direktivHTTPHandler struct {
	key      string
	pingAddr string

	mtx sync.Mutex

	requests map[string]*direktivHTTPRequest

	dbLog *dblog.Logger
}

func main() {

	d := &direktivHTTPHandler{
		requests: make(map[string]*direktivHTTPRequest),
	}

	if os.Getenv("DIREKTIV_DEBUG") == "true" {
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

	// prepare ping mechanism
	go d.pingMe()

	d.dbLog, err = setupLogging()
	if err != nil {
		log.Errorf("can not setup logging: %v", err)
	}

	// listen for sigterm
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		if d.dbLog != nil {
			d.dbLog.CloseConnection()
		}
	}()

	log.Infof("starting direktiv sidecar")
	log.Fatal(fasthttp.ListenAndServe(":8889", r.Handler))

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
			fmt.Sprintf("header incorrect: %s", err.Error()))
		return
	}

	// disable/enable ping
	if len(d.pingAddr) == 0 {
		d.pingAddr = vals[direktiv.DirektivPingAddrHeader]
	}

	log15log, err := d.dbLog.LoggerFunc("namespace", "instanceId")
	if err != nil {
		log.Errorf("can not setup logger: %v", err)
	}

	// add to request manager
	d.mtx.Lock()
	d.requests[vals[direktiv.DirektivActionIDHeader]] = &direktivHTTPRequest{
		ctx:    ctx,
		logger: log15log,
	}
	d.mtx.Unlock()

	defer func() {
		log.Debugf("cleanup request map")
		d.mtx.Lock()
		if _, ok := d.requests[vals[direktiv.DirektivActionIDHeader]]; ok {
			delete(d.requests, vals[direktiv.DirektivActionIDHeader])
		}
		d.mtx.Unlock()
	}()

	log.Infof("handle request %s", vals[direktiv.DirektivActionIDHeader])

	// FROM HERE WE NEED TO RESPOND VIA GRPC

	// forward request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		svcAddr, bytes.NewReader(ctx.Request.Body()))
	if err != nil {
		generateError(ctx, direktiv.ServiceErrorNetwork,
			fmt.Sprintf("create request failed: %s", err.Error()))
		return
	}

	// add header so client can use it ass reference
	req.Header.Add(direktiv.DirektivActionIDHeader,
		vals[direktiv.DirektivActionIDHeader])

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		generateError(ctx, direktiv.ServiceErrorNetwork,
			fmt.Sprintf("execute request failed: %s", err.Error()))
		return
	}

	defer resp.Body.Close()
	f, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		// TODO we need to respond to flow
		generateError(ctx, direktiv.ServiceErrorIO,
			fmt.Sprintf("read response body: %s", err.Error()))
		return
	}

	// respond to service
	respondToFlow(resp, vals[direktiv.DirektivInstanceIDHeader],
		vals[direktiv.DirektivActionIDHeader], int32(step), f)
	fmt.Fprintf(ctx, "ok")

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

func respondToFlow(resp *http.Response, iid, aid string, step int32, data []byte) {

	ec := direktiv.ServiceErrorImage
	em := ""

	r := &flow.ReportActionResultsRequest{
		InstanceId: &iid,
		Step:       &step,
		ActionId:   &aid,
		Output:     data,
	}

	// do we have error headers
	if resp.Header.Get(direktiv.DirektivErrorCodeHeader) != "" {
		ec = resp.Header.Get(direktiv.DirektivErrorCodeHeader)
		r.ErrorCode = &ec

		em = resp.Header.Get(direktiv.DirektivErrorMessageHeader)
		r.ErrorMessage = &em
	}

	// var resp direktiv.ServiceResponse
	// err := json.Unmarshal(data, &resp)
	// if err != nil {
	// 	// if error we can return internal errors
	// 	// because the container returned rubbish
	// 	r.ErrorCode = &ec
	//
	// 	em := err.Error()
	// 	r.ErrorMessage = &em
	// }

	// if the container reports an error we return that too
	// if len(resp.ErrorMessage) > 0 {
	// 	r.ErrorCode = &resp.ErrorCode
	// 	r.ErrorMessage = &resp.ErrorMessage
	// }

	// b, err := json.Marshal(resp.Data)

	// // we set output in any case
	// r.Output = b
	//
	// // but if it is not json we report an error
	// if err != nil {
	// 	r.ErrorCode = &ec
	// 	em := err.Error()
	// 	r.ErrorMessage = &em
	// }

	conn, err := grpc.Dial("localhost:7777", grpc.WithInsecure())
	if err != nil {
		log.Errorf("can not connect to flow: %v", err)
		return
	}
	flowClient := flow.NewDirektivFlowClient(conn)

	_, err = flowClient.ReportActionResults(context.Background(), r)
	if err != nil {
		log.Errorf("can not respond to flow: %v", err)
		return
	}

}
