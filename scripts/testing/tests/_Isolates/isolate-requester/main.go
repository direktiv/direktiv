package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/net/publicsuffix"

	b64 "encoding/base64"
)

const code = "com.%s.error"

// request the input object for the requester container
type request struct {
	Method             string                 `json:"method"`
	URL                string                 `json:"url"`
	Body               interface{}            `json:"body"`
	Headers            map[string]interface{} `json:"headers"`
	Params             map[string]interface{} `json:"params"`
	Username           string                 `json:"username"`
	Password           string                 `json:"password"`
	InsecureSkipVerify bool                   `json:"insecureSkipVerify"`
}

// output for the requester container
type output struct {
	Body       interface{} `json:"body,omitempty"` // when the response is able to be unmarshalled
	Headers    http.Header `json:"headers"`
	StatusCode int         `json:"status-code"`
	Status     string      `json:"status"`
}

func Log(msg string) {
	fmt.Println(msg)

	if logf == nil {
		return
	}

	fmt.Fprintln(logf, msg)
}

var hasFailed bool

func Error(code, msg string) {
	Log(fmt.Sprintf("ERROR: %s; %s", code, msg))

	if hasFailed {
		return
	}

	hasFailed = true

	m := map[string]string{
		"code": code,
		"msg":  msg,
	}

	data, _ := json.Marshal(m)

	err := ioutil.WriteFile("/direktiv-data/error.json", data, 0o755)
	if err != nil {
		Error("", err.Error())
	}
}

func Respond(data []byte) {
	err := ioutil.WriteFile("/direktiv-data/output.json", data, 0o755)
	if err != nil {
		Error("", err.Error())
	}
}

func Request(input []byte) {
	obj := new(request)
	err := json.Unmarshal(input, obj)
	if err != nil {
		Error(fmt.Sprintf(code, "unmarshal-input"), err.Error())
		return
	}

	var b []byte

	Log("Creating cookie jar")

	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		Error(fmt.Sprintf(code, "cookie-jar"), err.Error())
		return
	}

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: obj.InsecureSkipVerify,
			},
		},
	}

	Log("Creating new request")

	if obj.Body != nil {
		switch v := obj.Body.(type) {
		case string:
			Log("Body is a string ignore marshal.")
			b = []byte(obj.Body.(string))
		default:
			Log(fmt.Sprintf("Body is of type %v", v))
			b, err = json.Marshal(obj.Body)
			if err != nil {
				Error(fmt.Sprintf(code, "marshal-body"), err.Error())
				return
			}
		}

		Log("Body exists, attaching to the request")
	}

	Log("Creating URL...")
	u, err := url.Parse(obj.URL)
	if err != nil {
		Error(fmt.Sprintf(code, "url-parse"), err.Error())
		return
	}

	q := u.Query()
	for k, v := range obj.Params {
		var actualVal string
		// Handle other types provided and convert to string automatically
		switch t := v.(type) {
		case bool:
			actualVal = strconv.FormatBool(t)
		case float64:
			actualVal = strconv.FormatFloat(t, 'f', 6, 64)
		case string:
			actualVal = t
		}
		Log(fmt.Sprintf("Adding param %s=%s", k, actualVal))
		q.Set(k, actualVal)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(obj.Method, u.String(), bytes.NewReader(b))
	if err != nil {
		Error(fmt.Sprintf(code, "create-request"), err.Error())
		return
	}

	for k, v := range obj.Headers {
		var actualVal string
		// Handle other types provided and convert to string automatically
		switch t := v.(type) {
		case bool:
			actualVal = strconv.FormatBool(t)
		case float64:
			actualVal = strconv.FormatFloat(t, 'f', 6, 64)
		case string:
			actualVal = t
		}

		// Adding a header requires it to be a string so might as well sprintf
		req.Header.Add(k, actualVal)
	}

	if obj.Username != "" && obj.Password != "" {
		sEnc := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", obj.Username, obj.Password)))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", sEnc))
		Log("Adding Basic authorization header")
	}

	Log("Sending request...")

	resp, err := client.Do(req)
	if err != nil {
		Error(fmt.Sprintf(code, "send-request"), err.Error())
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Error(fmt.Sprintf(code, "read-resp-body"), err.Error())
		return
	}

	var mapBody map[string]interface{}
	var dataBody interface{}
	var responding output
	responding.Status = resp.Status
	responding.StatusCode = resp.StatusCode
	responding.Headers = resp.Header

	// if body is unable to be marshalled treat as a byte array
	err = json.Unmarshal(body, &mapBody)
	if err != nil {
		json.Unmarshal(body, &dataBody)
		responding.Body = dataBody
	} else {
		responding.Body = mapBody
	}

	data, err := json.Marshal(responding)
	if err != nil {
		Error(fmt.Sprintf(code, "marshal-output"), err.Error())
		return
	}

	Respond(data)
}

var logf *os.File

func initialize() error {
	var err error

	logf, err = os.Create("/direktiv-data/out.log")
	if err != nil {
		return err
	}

	return nil
}

func cleanup() {
	if logf != nil {
		logf.Close()
	}

	f, err := os.Create("/direktiv-data/done")
	if err == nil {
		f.Close()
	}
}

func main() {
	err := initialize()
	if err != nil {
		Error("", err.Error())
		return
	}

	defer cleanup()

	input, err := ioutil.ReadFile("/direktiv-data/input.json")
	if err != nil {
		Error("", err.Error())
		return
	}

	Log(fmt.Sprintf("INPUT: %s\n", input))

	Request(input)
}
