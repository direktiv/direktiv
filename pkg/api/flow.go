package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	protocol "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	prometheus "github.com/prometheus/client_golang/api"
)

type flowHandler struct {
	client     grpc.FlowClient
	prometheus prometheus.Client

	apiV2Address string
}

func newSingleHostReverseProxy(patchReq func(req *http.Request) *http.Request) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req = patchReq(req)
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}

func newFlowHandler(base *mux.Router, router *mux.Router, conf *core.Config) (*flowHandler, error) {
	flowAddr := fmt.Sprintf("localhost:%d", conf.GrpcPort)
	slog.Debug("Connecting to Direktiv flows.", "addr", flowAddr)

	flowConn, err := util.GetEndpointTLS(flowAddr)
	if err != nil {
		slog.Error("Failed to connect to Direktiv flows.", "addr", flowAddr, "error", err)
		return nil, err
	}
	slog.Info("Connected to Direktiv flows.", "addr", flowAddr)

	h := &flowHandler{
		client:       grpc.NewFlowClient(flowConn),
		apiV2Address: fmt.Sprintf("localhost:%d", conf.ApiV2Port),
	}

	prometheusAddr := fmt.Sprintf("http://%s", conf.Prometheus)
	slog.Debug("Connecting to Prometheus.", "addr", prometheusAddr)
	h.prometheus, err = prometheus.NewClient(prometheus.Config{Address: prometheusAddr})
	if err != nil {
		slog.Error("Failed to connect to Prometheus.", "addr", prometheusAddr, "error", err)
		return nil, err
	}
	slog.Info("Connected to Prometheus.", "addr", prometheusAddr)

	slog.Debug("Initializing API routes on the router.")
	h.initRoutes(router)
	slog.Debug("API routes have been successfully added to the router.")

	slog.Debug("Setting up reverse proxy handlers for API and namespace endpoints.")
	setupReverseProxyHandlers(base, router, h.apiV2Address)

	return h, nil
}

func setupReverseProxyHandlers(base *mux.Router, router *mux.Router, apiV2Address string) {
	proxy := newSingleHostReverseProxy(func(req *http.Request) *http.Request {
		req.Host = ""
		req.URL.Host = apiV2Address
		req.URL.Scheme = "http"

		return req
	})

	router.PathPrefix("/v2").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { proxy.ServeHTTP(w, r) }))
	base.PathPrefix("/ns").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { proxy.ServeHTTP(w, r) }))
}

func (h *flowHandler) initRoutes(r *mux.Router) {
	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-invoked Metrics workflowMetricsInvoked
	// ---
	// description: |
	//   Get metrics of invoked workflow instances.
	// summary: Gets Invoked Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-invoked", h.WorkflowMetricsInvoked)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-successful Metrics workflowMetricsSuccessful
	// ---
	// description: |
	//   Get metrics of a workflow, where the instance was successful.
	// summary: Gets Successful Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-successful", h.WorkflowMetricsSuccessful)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-failed Metrics workflowMetricsFailed
	// ---
	// description: |
	//   Get metrics of a workflow, where the instance failed.
	// summary: Gets Failed Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-failed", h.WorkflowMetricsFailed)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-failed Metrics workflowMetricsMilliseconds
	// ---
	// description: |
	//   Get the timing metrics of a workflow's instance.
	//   This returns a total sum of the milliseconds a workflow has been executed for.
	// summary: Gets Workflow Time Metrics
	// parameters:
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-milliseconds", h.WorkflowMetricsMilliseconds)

	// swagger:operation GET /api/namespaces/{namespace}/tree/{workflow}?op=metrics-state-milliseconds Metrics workflowMetricsStateMilliseconds
	// ---
	// description: |
	//   Get the state timing metrics of a workflow's instance.
	//   This returns the timing of individual states in a workflow.
	// summary: Gets a Workflow State Time Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: workflow
	//   type: string
	//   required: true
	//   description: 'path to target workflow'
	// responses:
	//   '200':
	//     "description": "successfully got workflow metrics"
	pathHandler(r, http.MethodGet, RN_GetWorkflowMetrics, "metrics-state-milliseconds", h.WorkflowMetricsStateMilliseconds)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/invoked Metrics namespaceMetricsInvoked
	// ---
	// description: |
	//   Get metrics of invoked workflows in the targeted namespace.
	// summary: Gets Namespace Invoked Workflow Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/invoked", h.NamespaceMetricsInvoked).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/successful Metrics namespaceMetricsSuccessful
	// ---
	// description: |
	//   Get metrics of successful workflows in the targeted namespace.
	// summary: Gets Namespace Successful Workflow Instances Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/successful", h.NamespaceMetricsSuccessful).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/failed Metrics namespaceMetricsFailed
	// ---
	// description: |
	//   Get metrics of failed workflows in the targeted namespace.
	// summary: Gets Namespace Failed Workflow Instances Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/failed", h.NamespaceMetricsFailed).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation GET /api/namespaces/{namespace}/metrics/milliseconds Metrics namespaceMetricsMilliseconds
	// ---
	// description: |
	//   Get timing metrics of workflows in the targeted namespace.
	// summary: Gets Namespace Workflow Timing Metrics
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got namespace metrics"
	r.HandleFunc("/namespaces/{ns}/metrics/milliseconds", h.NamespaceMetricsMilliseconds).Name(RN_GetNamespaceMetrics).Methods(http.MethodGet)

	// swagger:operation POST /api/namespaces/{namespace}/broadcast Other broadcastCloudevent
	// ---
	// description: |
	//   Broadcast a cloud event to a namespace.
	//   Cloud events posted to this api will be picked up by any workflows listening to the same event type on the namescape.
	//   The body of this request should follow the cloud event core specification defined at https://github.com/cloudevents/spec .
	// summary: Broadcast Cloud Event
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: body
	//   name: cloudevent
	//   required: true
	//   description: Cloud Event request to be sent.
	//   schema:
	//     type: object
	// responses:
	//   '200':
	//     "description": "successfully sent cloud event"
	r.HandleFunc("/namespaces/{ns}/broadcast", h.BroadcastCloudevent).Name(RN_NamespaceEvent).Methods(http.MethodPost)

	// swagger:operation GET /api/namespaces/{namespace}/event-listeners Events getEventListeners
	// ---
	// description: |
	//   Get current event listeners.
	// summary: Get current event listeners.
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got event listeners"
	handlerPair(r, RN_EventListeners, "/namespaces/{ns}/event-listeners", h.EventListeners, h.EventListenersSSE)

	// swagger:operation GET /api/namespaces/{namespace}/events Events getEventHistory
	// ---
	// description: |
	//   Get recent events history.
	// summary: Get events history.
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// responses:
	//   '200':
	//     "description": "successfully got events history"
	handlerPair(r, RN_EventHistory, "/namespaces/{ns}/events", h.EventHistory, h.EventHistorySSE)

	// swagger:operation POST /api/namespaces/{namespace}/events/{event}/replay Other replayCloudevent
	// ---
	// description: |
	//   Replay a cloud event to a namespace.
	// summary: Replay Cloud Event
	// parameters:
	// - in: path
	//   name: namespace
	//   type: string
	//   required: true
	//   description: 'target namespace'
	// - in: path
	//   name: event
	//   type: string
	//   required: true
	//   description: 'target cloudevent'
	// responses:
	//   '200':
	//     "description": "successfully replayed cloud event"
	r.HandleFunc("/namespaces/{ns}/events/{event:.*}/replay", h.ReplayEvent).Name(RN_NamespaceEvent).Methods(http.MethodPost)
}

func (h *flowHandler) EventListeners(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}
	namespace := mux.Vars(r)["ns"]

	in := &grpc.EventListenersRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventListeners(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) EventListenersSSE(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())
	namespace := mux.Vars(r)["ns"]

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.EventListenersRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventListenersStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)
	defer func() {
		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}
	}()

	go func() {
		defer close(ch)

		for {
			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x
		}
	}()

	sse(w, ch)
}

func (h *flowHandler) EventHistory(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}
	namespace := mux.Vars(r)["ns"]

	in := &grpc.EventHistoryRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventHistory(ctx, in)
	respond(w, resp, err)
}

func (h *flowHandler) EventHistorySSE(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())
	namespace := mux.Vars(r)["ns"]

	ctx := r.Context()
	p, err := pagination(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	in := &grpc.EventHistoryRequest{
		Pagination: p,
		Namespace:  namespace,
	}

	resp, err := h.client.EventHistoryStream(ctx, in)
	if err != nil {
		respond(w, resp, err)
		return
	}

	ch := make(chan interface{}, 1)
	defer func() {
		_ = resp.CloseSend()

		for {
			_, more := <-ch
			if !more {
				return
			}
		}
	}()

	go func() {
		defer close(ch)

		for {
			x, err := resp.Recv()
			if err != nil {
				ch <- err
				return
			}

			ch <- x
		}
	}()

	sse(w, ch)
}

// nolint:canonicalheader
func ToGRPCCloudEvents(r *http.Request) ([]cloudevents.Event, error) {
	var events []cloudevents.Event
	ct := r.Header.Get("Content-type")
	oct := ct

	// if batch mode we need to parse the body to multiple events
	if strings.HasPrefix(ct, "application/cloudevents-batch+json") {
		// load body
		data, err := loadRawBody(r)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &events)
		if err != nil {
			return nil, err
		}

		for i := range events {
			ev := events[i]
			if ev.ID() == "" {
				ev.SetID(uuid.New().String())
			}
			err = ev.Validate()
			if err != nil {
				return nil, err
			}
		}

		return events, nil
	}

	if strings.HasPrefix(ct, "application/json") {
		_, err := json.Marshal(r.Header)
		if err != nil {
			return nil, err
		}
		s := r.Header.Get("Ce-Type")
		if s == "" {
			ct = "application/cloudevents+json; charset=UTF-8"
			r.Header.Set("Content-Type", ct)
		}
	}

	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyData))

	msg := protocol.NewMessageFromHttpRequest(r)
	ev, err := binding.ToEvent(context.Background(), msg)
	if err != nil {
		goto generic
	}

	// validate:
	if ev.ID() == "" {
		ev.SetID(uuid.New().String())
	}
	err = ev.Validate()

	// azure hack for dataschema '#' which is an invalid cloudevent
	if err != nil && strings.HasPrefix(err.Error(), "dataschema: if present") {
		err = ev.Context.SetDataSchema("")
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		goto generic
	}

	events = append(events, *ev)

	return events, nil

generic:

	xerr := err
	unmarshalable := false

	m := make(map[string]interface{})

	if strings.HasPrefix(oct, "application/json") {
		err = json.Unmarshal(bodyData, &m)
		if err == nil {
			unmarshalable = true
		}
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	ev = &event

	uid := uuid.New()
	ev.SetID(uid.String())
	ev.SetType("noncompliant")
	ev.SetSource("unknown")
	ev.SetDataContentType(ct)
	if unmarshalable {
		err = ev.SetData(oct, m)
		if err != nil {
			return events, xerr
		}
	} else {
		err = ev.SetData(oct, bodyData)
		if err != nil {
			return events, xerr
		}
	}

	err = ev.Context.SetExtension("error", xerr.Error())
	if err != nil {
		return events, xerr
	}

	err = ev.Validate()
	if err != nil {
		return events, xerr
	}

	events = append(events, event)

	return events, nil
}

func (h *flowHandler) doBroadcast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]

	ces, err := ToGRPCCloudEvents(r)
	if err != nil {
		respond(w, nil, err)
		return
	}

	for i := range ces {
		d, err := json.Marshal(ces[i])
		if err != nil {
			respond(w, nil, err)
			return
		}

		in := &grpc.BroadcastCloudeventRequest{
			Namespace:  namespace,
			Cloudevent: d,
		}

		resp, err := h.client.BroadcastCloudevent(ctx, in)
		respond(w, resp, err)
	}
}

func (h *flowHandler) BroadcastCloudevent(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	h.doBroadcast(w, r)
}

func (h *flowHandler) ReplayEvent(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "this", this())

	ctx := r.Context()
	namespace := mux.Vars(r)["ns"]
	event := mux.Vars(r)["event"]

	in := &grpc.ReplayEventRequest{
		Namespace: namespace,
		Id:        event,
	}

	resp, err := h.client.ReplayEvent(ctx, in)
	respond(w, resp, err)
}
