package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func this() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}

func handlerPair(r *mux.Router, name, path string, handler, sseHandler func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(path, sseHandler).Name(name).Methods(http.MethodGet).Headers("Accept", "text/event-stream")
	r.HandleFunc(path, handler).Name(name).Methods(http.MethodGet)
}

func pathHandler(r *mux.Router, method, name, op string, handler func(http.ResponseWriter, *http.Request)) {

	root := "/namespaces/{ns}/tree"
	path := root + "/{path:.*}"

	r1 := r.HandleFunc(root, handler).Name(name).Methods(method)
	r2 := r.HandleFunc(path, handler).Name(name).Methods(method)

	if op != "" {
		r1.Queries("op", op)
		r2.Queries("op", op)
	}

}

func pathHandlerSSE(r *mux.Router, name, op string, handler func(http.ResponseWriter, *http.Request)) {

	root := "/namespaces/{ns}/tree"
	path := root + "/{path:.*}"

	r1 := r.HandleFunc(root, handler).Name(name).Methods(http.MethodGet).Headers("Accept", "text/event-stream")
	r2 := r.HandleFunc(path, handler).Name(name).Methods(http.MethodGet).Headers("Accept", "text/event-stream")

	if op != "" {
		r1.Queries("op", op)
		r2.Queries("op", op)
	}

}

func pathHandlerPair(r *mux.Router, name, op string, handler, sseHandler func(http.ResponseWriter, *http.Request)) {
	pathHandler(r, http.MethodGet, name, op, handler)
	pathHandlerSSE(r, name, op, handler)
}

func loadRawBody(r *http.Request) ([]byte, error) {

	limit := int64(1024 * 1024 * 32)

	if r.ContentLength > 0 {
		if r.ContentLength > limit {
			return nil, errors.New("request payload too large")
		}
		limit = r.ContentLength
	}

	rdr := io.LimitReader(r.Body, limit)

	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return nil, err
	}

	return data, nil

}

func getInt32(r *http.Request, key string) (int32, error) {

	s := r.URL.Query().Get(key)
	if s == "" {
		return 0, nil
	}

	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(n), nil

}

func pagination(r *http.Request) (*grpc.Pagination, error) {

	after := r.URL.Query().Get("after")

	first, err := getInt32(r, "first")
	if err != nil {
		return nil, err
	}

	before := r.URL.Query().Get("before")

	last, err := getInt32(r, "last")
	if err != nil {
		return nil, err
	}

	ofield := r.URL.Query().Get("order.field")
	direction := r.URL.Query().Get("order.direction")

	ffield := r.URL.Query().Get("filter.field")
	ftype := r.URL.Query().Get("filter.type")
	val := r.URL.Query().Get("filter.val")

	p := &grpc.Pagination{
		After:  after,
		First:  first,
		Before: before,
		Last:   last,
		Order: &grpc.PageOrder{
			Field:     ofield,
			Direction: direction,
		},
		Filter: &grpc.PageFilter{
			Field: ffield,
			Type:  ftype,
			Val:   val,
		},
	}

	return p, nil

}

func badRequest(w http.ResponseWriter, err error) {
	code := http.StatusBadRequest
	msg := http.StatusText(code)
	http.Error(w, msg, code)
	return
}

func respond(w http.ResponseWriter, resp interface{}, err error) {

	if err != nil {

		code := ConvertGRPCStatusCodeToHTTPCode(status.Code(err))

		msg := http.StatusText(code)
		http.Error(w, msg, code)
		return
	}

	if resp == nil {
		goto nodata
	}

	if _, ok := resp.(*emptypb.Empty); ok {
		goto nodata
	}

	marshal(w, resp, true)

nodata:

	return

}

func marshal(w io.Writer, x interface{}, multiline bool) {

	data, err := protojson.MarshalOptions{
		Multiline:       multiline,
		EmitUnpopulated: true,
	}.Marshal(x.(proto.Message))
	if err != nil {
		panic(err)
	}

	s := string(data)

	fmt.Fprintf(w, "%s", s)

}

func unmarshalBody(r *http.Request, x interface{}) error {

	limit := int64(1024 * 1024 * 32)

	if r.ContentLength > 0 {
		if r.ContentLength > limit {
			return errors.New("request payload too large")
		}
		limit = r.ContentLength
	}

	dec := json.NewDecoder(io.LimitReader(r.Body, limit))
	dec.DisallowUnknownFields()

	err := dec.Decode(x)
	if err != nil {
		return err
	}

	return nil

}

func pathAndRef(r *http.Request) (string, string) {

	path, _ := mux.Vars(r)["path"]
	ref := r.URL.Query().Get("ref")
	return path, ref

}

// SSE Util functions

func sse(w http.ResponseWriter, ch <-chan interface{}) {

	flusher, err := sseSetup(w)
	if err != nil {
		return
	}

	for {

		select {

		case x, more := <-ch:

			if !more {
				return
			}

			var ok bool
			err, ok = x.(error)
			if ok {
				sseError(w, flusher, err)
				return
			}

			err = sseWriteJSON(w, flusher, x)
			if err != nil {
				return
			}

		case <-time.After(time.Second * 10):
			err = sseHeartbeat(w, flusher)
			if err != nil {
				return
			}

		}

	}

}

func sseError(w http.ResponseWriter, flusher http.Flusher, err error) {

	eo := GenerateErrObject(err)

	b, err := json.Marshal(eo)
	if err != nil {
		panic(err)
	}

	_, err = w.Write([]byte(fmt.Sprintf("event: error\ndata: %s\n\n", string(b))))
	if err != nil {
		return
	}

	flusher.Flush()

}

func sseSetup(w http.ResponseWriter) (http.Flusher, error) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return flusher, fmt.Errorf("streaming unsupported")
	}

	return flusher, nil

}

func sseWriteJSON(w http.ResponseWriter, flusher http.Flusher, data interface{}) error {

	buf := new(bytes.Buffer)

	marshal(buf, data, false)

	return sseWrite(w, flusher, buf.Bytes())

}

func sseWrite(w http.ResponseWriter, flusher http.Flusher, data []byte) error {

	_, err := w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(data))))
	if err != nil {
		return err
	}

	flusher.Flush()
	return nil

}

func sseHeartbeat(w http.ResponseWriter, flusher http.Flusher) error {

	_, err := w.Write([]byte(fmt.Sprintf("data: %s\n\n", "")))
	if err != nil {
		return err
	}

	flusher.Flush()
	return nil

}
