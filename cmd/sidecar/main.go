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
	"sync"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/vorteil/direktiv/pkg/direktiv"
	dlog "github.com/vorteil/direktiv/pkg/dlog"
	dblog "github.com/vorteil/direktiv/pkg/dlog/db"
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

func (d *direktivHTTPHandler) handleLog(aid, data string) {
	// d.requests[vals[direktiv.DirektivActionIDHeader]].logger
	log.Infof("%s", data)
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

	// ! Namespace  *string           `protobuf:"bytes,1,opt,name=namespace,proto3,oneof" json:"namespace,omitempty"`
	// ! InstanceId *string           `protobuf:"bytes,2,opt,name=instanceId,proto3,oneof" json:"instanceId,omitempty"`
	// ! Step       *int32            `protobuf:"varint,3,opt,name=step,proto3,oneof" json:"step,omitempty"`
	// ! Timeout    *int64            `protobuf:"varint,4,opt,name=timeout,proto3,oneof" json:"timeout,omitempty"`
	// ! ActionId   *string           `protobuf:"bytes,5,opt,name=actionId,proto3,oneof" json:"actionId,omitempty"`
	// Image      *string           `protobuf:"bytes,6,opt,name=image,proto3,oneof" json:"image,omitempty"`
	// Command    *string           `protobuf:"bytes,7,opt,name=command,proto3,oneof" json:"command,omitempty"`
	// ! Data       []byte            `protobuf:"bytes,8,opt,name=data,proto3,oneof" json:"data,omitempty"`
	// Registries map[string]string `protobuf:"bytes,9,rep,name=registries,proto3" json:"registries,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Size       *int32            `protobuf:"varint,10,opt,name=size,proto3,oneof" json:"size,omitempty"`

	// headers to check for
	hdrs := []string{direktiv.DirektivExchangeKeyHeader,
		direktiv.DirektivActionIDHeader,
		direktiv.DirektivPingAddrHeader}

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

	// setup local ping
	if len(d.pingAddr) == 0 {
		d.pingAddr = vals[direktiv.DirektivPingAddrHeader]
	}

	// disable ping
	if d.pingAddr == "noping" {
		d.pingAddr = ""
	}

	log15log, err := d.dbLog.LoggerFunc("namespace", "instanceId")
	if err != nil {
		log.Errorf("can not setup logger: %v", err)
	}

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
		generateError(ctx, direktiv.ServiceErrorIO,
			fmt.Sprintf("read response body: %s", err.Error()))
		return
	}

	// respond to service
	fmt.Println("RESPOND WITH DATA!!!")
	fmt.Println(string(f))

	fmt.Fprintf(ctx, string(f))

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
