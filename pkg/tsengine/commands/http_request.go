package commands

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dario.cat/mergo"
	"github.com/direktiv/direktiv/pkg/tsengine/runtime"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/hashicorp/go-retryablehttp"
)

type RequestCommand struct {
	rt *runtime.Runtime
}

func NewRequestCommand(rt *runtime.Runtime) *RequestCommand {
	return &RequestCommand{
		rt: rt,
	}
}

func (rc RequestCommand) GetName() string {
	return "httpRequest"
}

func (rc RequestCommand) GetCommandFunction() interface{} {
	return rc.HttpRequest
}

type Retry struct {
	Count int
	Wait  int
}

type HttpArgs struct {
	Method string
	URL    string
	Header map[string]string

	SkipSecure bool
	Timeout    int

	Input  interface{}
	File   *File
	Async  bool
	Retry  Retry
	Result string
}

const (
	httpResultJSON   = "json"
	httpResultString = "string"
	httpResultFile   = "file"
)

func (rc RequestCommand) HttpRequest(in interface{}) interface{} {

	args, err := utils.DoubleMarshal[HttpArgs](in)
	if err != nil {
		runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
	}

	defaultArgs := &HttpArgs{
		Retry: Retry{
			Count: 0,
			Wait:  5,
		},
		Method:  "POST",
		Result:  httpResultJSON,
		Timeout: 5,
	}

	// merge defaults in
	err = mergo.Merge(&args, defaultArgs)
	if err != nil {
		runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
	}

	// fail if there is no url
	if args.URL == "" {
		runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, fmt.Errorf("url can not be undefined"))
	}

	var rd io.Reader
	if args.File != nil {
		rd, err = os.Open(args.File.RealPath)
		if err != nil {
			runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
		}
		// that is guaranteed cast-able
		defer rd.(*os.File).Close()
	} else if args.Input != "" {
		switch v := args.Input.(type) {
		case string:
			rd = strings.NewReader(v)
		case nil:
		default:
			b, err := json.Marshal(v)
			if err != nil {
				runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
			}
			rd = strings.NewReader(string(b))
		}
	}

	req, err := http.NewRequest(args.Method, args.URL, rd)
	if err != nil {
		runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
	}

	if len(args.Header) > 0 {
		for k, v := range args.Header {
			req.Header.Add(k, v)
		}
	}

	retryClient := retryablehttp.NewClient()
	// TODO: logger
	// retryClient.Logger =
	retryClient.HTTPClient.Timeout = time.Duration(args.Timeout) * time.Second

	if args.SkipSecure {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		retryClient.HTTPClient.Transport = transport
	}

	retryClient.Backoff = retryablehttp.DefaultBackoff
	retryClient.RetryMax = args.Retry.Count
	retryClient.RetryWaitMax = time.Duration(args.Retry.Wait*3) * time.Second
	retryClient.RetryWaitMin = time.Duration(args.Retry.Wait) * time.Second

	// retry request
	rreq, err := retryablehttp.FromRequest(req)
	if err != nil {
		runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
	}

	if args.Async {
		go rc.doRequest(retryClient, rreq, args.Result)
		runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
	}

	data, err := rc.doRequest(retryClient, rreq, args.Result)
	if err != nil {
		runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivHTTPErrorCode, err)
	}

	return data
}

func (rc RequestCommand) doRequest(client *retryablehttp.Client, req *retryablehttp.Request, result string) (interface{}, error) {
	resp, err := client.Do(req)
	if err != nil {
		// easier option than .Is unwrapping to urlError
		if strings.Contains(err.Error(), "context deadline exceeded") {
			runtime.ThrowRuntimeError(rc.rt.VM, runtime.DirektivTimeoutErrorCode, err)
		}
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp != nil && resp.Header.Get(runtime.DirektivErrorCodeHeader) != "" {
		code := resp.Header.Get(runtime.DirektivErrorCodeHeader)
		msg := resp.Header.Get(runtime.DirektivErrorMessageHeader)
		runtime.ThrowRuntimeError(rc.rt.VM, code, fmt.Errorf("%s", msg))
	}

	// switch result
	switch result {
	case httpResultFile:
		outPath := filepath.Join(rc.rt.DirInfo().InstanceDir, "result.data")
		err := writeFile(outPath, resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, nil
	case httpResultJSON, httpResultString:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// try to return JSON, string otherwise
		if result == httpResultJSON && json.Valid(b) {
			var out interface{}
			err = json.Unmarshal(b, &out)
			return out, err
		} else {
			return string(b), nil
		}
	default:
		return nil, fmt.Errorf("unknown result type %s", result)
	}

}

func writeFile(outPath string, src io.ReadCloser) error {

	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(f, src)

	return src.Close()
}
