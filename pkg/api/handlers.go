package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/itchyny/gojq"
)

const (
	// GRPCCommandTimeout : timeout for grpc function calls
	GRPCCommandTimeout = 90 * time.Second
)

type Handler struct {
	s *Server
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
	/* #nosec */
	_, _ = w.Write([]byte(strings.Join(jqResults, "\n")))
}
