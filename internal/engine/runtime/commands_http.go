package runtime

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/sobek"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type httpRequestObject struct {
	Method string

	Body    any
	Headers map[string]string
	Params  map[string]string

	SkipTls bool

	Username, Password string
	Timeout            int
}

type httpResponseObject struct {
	responseType string

	url string

	redirected bool

	status int

	statusText string

	ok bool

	headers http.Header

	body []byte

	err string
}

func doHttpRequest(addr string, config any) (*httpResponseObject, error) {
	// url requires value
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("url invalid for request")
	}

	// map sure it is a map
	if config != nil {
		_, ok := config.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("request configuration has invalid value")
		}
	}

	// double marshal
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	var req httpRequestObject
	err = json.Unmarshal(b, &req)
	if err != nil {
		return nil, err
	}

	// if config is empty, set the method to get
	if req.Method == "" {
		req.Method = http.MethodGet
	}

	// from here we can always respond with an object
	obj := &httpResponseObject{
		url:          u.String(),
		responseType: "error",
		redirected:   false,
		status:       0,
		statusText:   "",
		ok:           false,
		headers:      http.Header{},
	}

	// generate query
	q := u.Query()
	for k, v := range req.Params {
		q.Add(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = q.Encode()

	var rBody io.Reader = nil
	if req.Body != nil {
		b, _ = json.Marshal(req.Body)
		rBody = strings.NewReader(string(b))
	}

	request, err := http.NewRequest(req.Method, u.String(), rBody)
	if err != nil {
		obj.err = err.Error()
		return obj, nil
	}

	// add headers
	for k, v := range req.Headers {
		request.Header.Add(k, v)
	}

	if req.Username != "" && req.Password != "" {
		sEnc := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", req.Username, req.Password)))
		request.Header.Add("Authorization", fmt.Sprintf("Basic %s", sEnc))
	}

	// with default timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// set timeout if configured
	if req.Timeout > 0 {
		client.Timeout = time.Duration(req.Timeout) * time.Second
	}

	// skip tls if configured
	if req.SkipTls {
		cr := http.DefaultTransport.(*http.Transport).Clone()
		cr.TLSClientConfig = &tls.Config{InsecureSkipVerify: req.SkipTls}
		client.Transport = cr
	}

	resp, err := client.Do(request)
	if err != nil {
		obj.err = err.Error()
		return obj, nil
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		obj.err = err.Error()
		return obj, nil
	}
	obj.body = body

	isOK := resp.StatusCode >= 200 && resp.StatusCode <= 299

	obj.ok = isOK

	obj.headers = resp.Header
	obj.responseType = "basic"
	obj.status = resp.StatusCode
	obj.statusText = resp.Status

	return obj, nil
}

func (rt *Runtime) populateResponseObject(response *httpResponseObject) *sobek.Object {
	responseObject := rt.vm.NewObject()
	responseObject.Set("responseType", response.responseType)
	responseObject.Set("error", response.err)
	responseObject.Set("ok", response.ok)
	responseObject.Set("redirected", response.redirected)
	responseObject.Set("status", response.status)
	responseObject.Set("statusText", response.statusText)
	responseObject.Set("url", response.url)

	m := rt.vm.NewObject()
	for k, vs := range response.headers {
		arr := rt.vm.NewArray()
		for i, v := range vs {
			arr.Set(strconv.Itoa(i), v)
		}
		m.Set(k, arr)
	}
	responseObject.Set("headers", m)

	responseObject.Set("text", func(call sobek.FunctionCall) sobek.Value {
		return rt.vm.ToValue(string(response.body))
	})
	responseObject.Set("json", func(call sobek.FunctionCall) sobek.Value {
		var r any
		err := json.Unmarshal(response.body, &r)
		if err != nil {
			panic(rt.vm.ToValue(err.Error()))
		}

		return rt.vm.ToValue(r)
	})

	return responseObject
}

func (rt *Runtime) fetchSync(addr string, config any) *sobek.Object {
	span := rt.tracingPack.trace("calling http")
	span.SetAttributes(attribute.KeyValue{
		Key:   "url",
		Value: attribute.StringValue(addr),
	})
	defer span.End()

	response, err := doHttpRequest(addr, config)
	if err != nil {
		rt.tracingPack.thrownError = err
		span.SetStatus(codes.Error, err.Error())
		panic(rt.vm.ToValue(err.Error()))
	}

	span.SetStatus(codes.Ok, codes.Ok.String())
	return rt.populateResponseObject(response)
}

func (rt *Runtime) fetch(addr string, config any) *sobek.Promise {
	span := rt.tracingPack.trace("calling http async")
	defer span.End()
	p, resolve, reject := rt.vm.NewPromise()
	go func() {
		response, err := doHttpRequest(addr, config)
		if err != nil {
			reject(rt.vm.ToValue(err.Error()))
			return
		}
		resolve(rt.populateResponseObject(response))
	}()

	return p
}
