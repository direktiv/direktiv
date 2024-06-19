package state

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/direktiv/direktiv/pkg/tsengine/commands"
	"github.com/direktiv/direktiv/pkg/tsengine/runtime"
)

type State struct {
	data interface{}

	header http.Header
	params url.Values

	response *Response
}

func NewState(res http.ResponseWriter, data interface{},
	headers http.Header, params url.Values, runtime *runtime.Runtime) *State {

	state := &State{
		response: NewResponse(res, runtime),
		data:     data,
		header:   headers,
		params:   params,
	}

	return state
}

func (s *State) Data() interface{} {
	return s.data
}

func (s *State) GetHeader(name string) string {
	return s.header.Get(name)
}

func (s *State) GetHeaderValues(name string) []string {
	return s.header.Values(name)
}

func (s *State) GetParam(name string) string {
	return s.params.Get(name)
}

func (s *State) GetParamValues(name string) []string {
	return s.params[name]
}

func (s *State) Response() *Response {
	return s.response
}

type Response struct {
	runtime  *runtime.Runtime
	response http.ResponseWriter
	Written  bool
}

func NewResponse(resp http.ResponseWriter, runtime *runtime.Runtime) *Response {

	// we set a dummy if there is no response
	// e.g. cron, restart or event
	if resp == nil {
		resp = &DummyResponse{}
	}

	return &Response{
		response: resp,
		runtime:  runtime,
	}
}

func (r *Response) AddHeader(key, value string) {
	r.Written = true
	r.response.Header().Add(key, value)
}

func (r *Response) SetHeader(key, value string) {
	r.Written = true
	r.response.Header().Set(key, value)
}

func (r *Response) SetStatus(status int) {
	r.Written = true
	r.response.WriteHeader(status)
}

func (r *Response) Write(data string) {
	r.Written = true
	_, err := r.response.Write([]byte(data))

	if err != nil {
		runtime.ThrowRuntimeError(r.runtime.VM, runtime.DirektivFileErrorCode, err)
	}
}

func (r *Response) WriteFile(f *commands.File) {
	r.Written = true
	file, err := os.OpenFile(f.RealPath, os.O_RDONLY, 0400)
	if err != nil {
		runtime.ThrowRuntimeError(r.runtime.VM, runtime.DirektivFileErrorCode, err)
	}
	defer file.Close()

	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			runtime.ThrowRuntimeError(r.runtime.VM, runtime.DirektivFileErrorCode, err)
		}
		if n == 0 {
			break
		}
		if _, err := r.response.Write(buf[:n]); err != nil {
			runtime.ThrowRuntimeError(r.runtime.VM, runtime.DirektivFileErrorCode, err)
		}
	}
}

type DummyResponse struct {
}

func (d *DummyResponse) Header() http.Header {
	return http.Header{}
}

func (d *DummyResponse) Write(b []byte) (int, error) {
	return len(b), nil
}

func (d *DummyResponse) WriteHeader(statusCode int) {
}
