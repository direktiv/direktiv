package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func handlerPair(r *mux.Router, name, path string, handler, sseHandler func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(path, handler).Name(name).Methods(http.MethodGet)
	r.HandleFunc(path, sseHandler).Name(name).Methods(http.MethodGet).Headers("Accept", "text/event-stream")
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

	marshal(w, resp)

nodata:

	w.WriteHeader(http.StatusNoContent)
	return

}

func marshal(w http.ResponseWriter, x interface{}) {

	data, err := protojson.MarshalOptions{
		Multiline:       true,
		EmitUnpopulated: true,
	}.Marshal(x.(proto.Message))
	if err != nil {
		panic(err)
	}

	s := string(data)

	fmt.Fprintf(w, "%s", s)

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

	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return sseWrite(w, flusher, b)

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
