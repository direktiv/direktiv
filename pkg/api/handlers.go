package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/itchyny/gojq"
	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// GRPCCommandTimeout : timeout for grpc function calls
	GRPCCommandTimeout = 30 * time.Second
)

type Handler struct {
	s *Server
}

func (h *Handler) workflowMetrics(w http.ResponseWriter, r *http.Request) {

	var err error
	ns := mux.Vars(r)["namespace"]
	wf := mux.Vars(r)["workflow"]

	// QueryParams
	values := r.URL.Query()
	since := values.Get("since")

	var x time.Time
	if since != "" {
		dura, err := time.ParseDuration(since)
		if err != nil {
			ErrResponse(w, err)
			return
		}
		x = time.Now().Add(-1 * dura)
	}

	ts := timestamppb.New(x)

	in := &ingress.WorkflowMetricsRequest{
		Namespace:      &ns,
		Workflow:       &wf,
		SinceTimestamp: ts,
	}

	// GRPC Context
	gCTX := r.Context()
	gCTX, cancel := context.WithDeadline(gCTX, time.Now().Add(GRPCCommandTimeout))
	defer cancel()

	resp, err := h.s.direktiv.WorkflowMetrics(gCTX, in)
	if err != nil {
		// Convert error
		s := status.Convert(err)
		ErrResponse(w, fmt.Errorf(s.Message()))
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := h.s.json.Marshal(w, resp); err != nil {
		ErrResponse(w, err)
		return
	}

}

func (h *Handler) jqPlayground(w http.ResponseWriter, r *http.Request) {

	var jqBody JQQuery
	// Read Body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	err = json.Unmarshal(b, &jqBody)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	query, err := gojq.Parse(jqBody.Query)
	if err != nil {
		ErrResponse(w, fmt.Errorf("jq Filter is invalid: %v", err))
		return
	}

	jqResults := make([]string, 0)
	iter := query.Run(jqBody.Input) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			ErrResponse(w, err)
			return
		}

		b, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			ErrResponse(w, err)
			return
		}

		jqResults = append(jqResults, string(b))
	}

	// Write Response
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strings.Join(jqResults, "\n")))
}
