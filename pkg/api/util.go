package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
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
	pathHandlerSSE(r, name, op, sseHandler)
	pathHandler(r, http.MethodGet, name, op, handler)
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

	data, err := io.ReadAll(rdr)
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
	limit, err := getInt32(r, "limit")
	if err != nil {
		return nil, err
	}

	offset, err := getInt32(r, "offset")
	if err != nil {
		return nil, err
	}

	var l1, l2, l3 int
	var ok bool
	var orderfields []string
	var orderdirection []string
	var orderings []*grpc.PageOrder
	if orderfields, ok = r.URL.Query()["order.field"]; ok {
		l1 = len(orderfields)
	}
	if orderdirection, ok = r.URL.Query()["order.direction"]; ok {
		l2 = len(orderdirection)
	}
	if l1 == 1 && l2 == 0 {
		ofield := orderfields[0]
		orderings = append(orderings, &grpc.PageOrder{
			Field: ofield,
		})
	} else {
		if l1 != l2 {
			return nil, errors.New("bad ordering arguments")
		}
		if l1 > 0 {
			for i := range orderfields {
				ofield := orderfields[i]
				direction := orderdirection[i]
				orderings = append(orderings, &grpc.PageOrder{
					Field:     ofield,
					Direction: direction,
				})
			}
		}
	}

	l1 = 0
	l2 = 0
	var filterfields []string
	var filtertypes []string
	var filtervals []string
	var filters []*grpc.PageFilter
	if filterfields, ok = r.URL.Query()["filter.field"]; ok {
		l1 = len(filterfields)
	}
	if filtertypes, ok = r.URL.Query()["filter.type"]; ok {
		l2 = len(filtertypes)
	}
	if filtervals, ok = r.URL.Query()["filter.val"]; ok {
		l3 = len(filtervals)
	}
	if l1 != l2 || l1 != l3 {
		return nil, errors.New("bad filtering arguments")
	}
	if l1 > 0 {
		for i := range filterfields {
			ffield := filterfields[i]
			ftype := filtertypes[i]
			fval := filtervals[i]
			filters = append(filters, &grpc.PageFilter{
				Field: ffield,
				Type:  ftype,
				Val:   fval,
			})
		}
	}

	p := &grpc.Pagination{
		Limit:  limit,
		Offset: offset,
		Order:  orderings,
		Filter: filters,
	}

	return p, nil
}

func badRequest(w http.ResponseWriter, err error) {
	code := http.StatusBadRequest
	msg := http.StatusText(code)
	http.Error(w, msg, code)
}

func respondStruct(w http.ResponseWriter, resp interface{}, code int, err error) {
	w.WriteHeader(code)

	if err != nil {
		logger.Errorf("grpc error: %v", err.Error())
		msg := http.StatusText(code)
		http.Error(w, msg, code)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// OkBody is an arbitrary placeholder response that represents an ok response body
//
// swagger:model
type OkBody map[string]interface{}

// swagger:model
type ErrorResponse interface {
	// swagger:name Message
	Error() string
	// swagger:name StatusCode
	StatusCode() int
}

// swagger: model ErrorBack
type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ErrorBody) StatusCode() int {
	return e.Code
}

func respond(w http.ResponseWriter, resp interface{}, err error) {
	if err != nil {
		code := ConvertGRPCStatusCodeToHTTPCode(status.Code(err))
		st := status.Convert(err)

		o := &ErrorBody{
			Code:    code,
			Message: st.Message(),
		}

		data, err := json.Marshal(&o)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)

		_, _ = io.Copy(w, bytes.NewReader(data))

		return
	}

	if resp == nil {
		goto nodata
	}

	if _, ok := resp.(*emptypb.Empty); ok {
		goto nodata
	}

	w.Header().Set("Content-Type", "application/json")
	marshal(w, resp, true)

nodata:

	return
}

func respondJSON(w http.ResponseWriter, resp interface{}, err error) {
	if err != nil {
		code := ConvertGRPCStatusCodeToHTTPCode(status.Code(err))

		msg := err.Error()
		http.Error(w, msg, code)
		return
	}

	if resp == nil {
		goto nodata
	}

	if _, ok := resp.(*emptypb.Empty); ok {
		goto nodata
	}

	w.Header().Set("Content-Type", "application/json")
	marshalJSON(w, resp, true)

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

func marshalJSON(w io.Writer, x interface{}, multiline bool) {
	data, err := json.Marshal(x)
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
	path := mux.Vars(r)["path"]
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

		case <-time.After(time.Second * 5):
			err = sseHeartbeat(w, flusher)
			if err != nil {
				return
			}
		}
	}
}

func sseOnce(w http.ResponseWriter, ch <-chan interface{}) {
	x, more := <-ch
	if !more {
		return
	}

	err, ok := x.(error)
	if ok {
		sseError(w, nil, err)
		return
	}

	err = sseWriteJSON(w, nil, x)
	if err != nil {
		return
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

	if flusher != nil {
		flusher.Flush()
	}
}

func sseSetup(w http.ResponseWriter) (http.Flusher, error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

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
	_, err := io.Copy(w, strings.NewReader(fmt.Sprintf("data: %s\n\n", string(data))))
	if err != nil {
		return err
	}

	if flusher != nil {
		flusher.Flush()
	}

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

// Swagger Param Wrappers
// IMPORTANT: HOW TO LINK PARAMETERS TO OPERATIONS
// You can link parameters from a struct to a operations by adding the operation id to it.
// e.g. `swagger:parameters getWorkflowLogs` will add struct parameters to the getWorkflowLogs operation
//
// Once you've linked a parameter it will automatically merge with the target operation parameters starting from the top with the exported fields from the struct
// To deal with this we can add dummy parameters to like so (ref: https://github.com/go-swagger/go-swagger/issues/1416):
// parameters:
// - "": "#/parameters/PaginationQuery/order.field"
// The order is very important. For example if our struct is exporting the fields: order.field, order.direction, filter.field, filter.type
// We need to setup the operations parameters like so:
// parameters:
// - "": "#/parameters/PaginationQuery/order.field"
// - "": "#/parameters/PaginationQuery/order.direction"
// - "": "#/parameters/PaginationQuery/filter.field"
// - "": "#/parameters/PaginationQuery/filter.type"
// - Any other parameters you wana define
//
// Note: dummy parameters must be the first parameters defined in a operation
// Note: Because swagger:parameters are merged into the operation dummy parameters, we can do useful stuff like this in the operation dummy parameters
// parameters:
// - "": "#/parameters/PaginationQuery/order.field"
//   enum:
//     - CREATED
//     - UPDATED
// ....
// This can be useful for when different operations have different fields they can order on.

// swagger:parameters getWorkflowLogs getNamespaces serverLogs namespaceLogs instanceLogs getInstanceList getWorkflowLogs
type PaginationQuery struct {
	// field to order by
	//
	// in: query
	// name: "order.field"
	// type: string
	// required: false
	// description: "field to order by"
	PageOrderField string `json:"order.field"`

	// order direction
	//
	// in: query
	// name: "order.direction"
	// type: string
	// required: false
	// description: "order direction"
	// enum: DESC, ASC
	PageOrderDirection string `json:"order.direction"`

	// field to filter
	//
	// in: query
	// name: "filter.field"
	// type: string
	// required: false
	// description: "field to filter"
	PageFilterField string `json:"filter.field"`

	// filter behaviour
	//
	// in: query
	// name: "filter.type"
	// type: string
	// required: false
	// description: "filter behaviour"
	PageFilterType string `json:"filter.type"`
}

type telemetryHandler struct {
	srv  *Server
	next http.Handler
}

func (h *telemetryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	span := trace.SpanFromContext(ctx)
	tid := span.SpanContext().TraceID()

	var annotations []interface{}
	annotations = append(annotations, "trace", tid.String())

	v := mux.Vars(r)

	if s, exists := v["ns"]; exists {
		annotations = append(annotations, "namespace", s)
	}

	if s, exists := v["instance"]; exists {
		annotations = append(annotations, "instance", s)
	}

	if s, exists := v["path"]; exists {
		annotations = append(annotations, "workflow", GetInodePath(s))
	}

	if s, exists := v["var"]; exists {
		annotations = append(annotations, "variable", s)
	}

	if s, exists := v["secret"]; exists {
		annotations = append(annotations, "secret", s)
	}

	if s, exists := v["svn"]; exists {
		annotations = append(annotations, "service", s)
	}

	if s, exists := v["rev"]; exists {
		annotations = append(annotations, "servicerevision", s)
	}

	if s, exists := v["pod"]; exists {
		annotations = append(annotations, "pod", s)
	}

	if s := r.URL.Query().Get("op"); s != "" {
		annotations = append(annotations, "pathoperation", s)
	}

	annotations = append(annotations, "routename", mux.CurrentRoute(r).GetName())
	annotations = append(annotations, "httpmethod", r.Method)
	annotations = append(annotations, "httppath", r.URL.Path)

	// response
	// token

	h.srv.logger.Debugw("Handling request", annotations...)

	h.next.ServeHTTP(w, r)
}

func (s *Server) logMiddleware(h http.Handler) http.Handler {
	return &telemetryHandler{
		srv:  s,
		next: h,
	}
}
