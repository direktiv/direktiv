package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"bytes"

	"github.com/fasthttp/router"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type direktivHTTPHandler struct {
	key string
}

func main() {

d := &direktivHTTPHandler{}

	if os.Getenv("DIREKTIV_DEBUG") == "true" {
		log.SetLevel(logrus.DebugLevel)
	}

	k, err := ioutil.ReadFile("/var/secret/exchangeKey")
	if (err != nil) {
		log.Errorf("can not read exchange key: %v", err)
	}

	// store the key
	d.key = string(k)

	r := router.New()
	r.POST("/", d.Base)
	r.GET("/ping", d.Ping)

	log.Infof("starting direktiv request handler")
	log.Fatal(fasthttp.ListenAndServe(":8889", r.Handler))

}

func  (d *direktivHTTPHandler) Ping(ctx *fasthttp.RequestCtx) {
	log.Debugf("direktiv sidecar alive ping")
	ctx.WriteString("pong")
}

// Base is the main function receiving requests and handling pings/logs and
// response if required
func  (d *direktivHTTPHandler) Base(ctx *fasthttp.RequestCtx) {

	// check if secret is available, otherwise decline request
	// check if d.Key is in Header

	// count request, if not nil keep pinging

	// wait for cancel request, end ctx for request (pass-through)


	log.Infof("CONTENT HEADER %v", ctx.Request.Header.ContentLength());
	log.Infof("KEY IS %v", d.key)


	// req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", ctx.RequestBodyStream())
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", bytes.NewReader(ctx.Request.Body()))
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	resp.Body.Close()
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}

	fmt.Println("THIS IS THE DATA!!!!")
	fmt.Println(string(f))

	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value") // This makes it work

	fmt.Fprintf(ctx, string(f))


}
