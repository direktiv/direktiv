package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HandleFunc func(http.ResponseWriter, *http.Request)
type WrapperFunc func(*http.Request) *http.Request

// Handler ..
type Handler struct {
	onSuccess func() error
	onFailure func() error
	exec      func(http.ResponseWriter, *http.Request)
}

func NewHandler(fn HandleFunc) *Handler {
	return &Handler{
		exec: fn,
	}
}

const (
	CtxWrapperError      = "CTX_WRAPPER_ERROR"
	CtxWrapperStatusCode = "CTX_WRAPPER_STATUS_CODE"
)

func checkWrapperError(w http.ResponseWriter, r *http.Request) bool {
	// returns true if error was detected
	// and automatically responds with msg/code if true

	x := r.Context().Value(CtxWrapperError)
	if x != nil {
		// wrapper set a non-nil error message
		err, _ := x.(error)
		if err != nil {
			err = fmt.Errorf("unknown error")
		}

		// get response status code from context
		var code = http.StatusInternalServerError
		y := r.Context().Value(CtxWrapperStatusCode)
		if y != nil {
			c, _ := y.(int)
			if c != 0 {
				code = c
			}
		}

		w.WriteHeader(code)
		io.Copy(w, strings.NewReader(err.Error()))
		return true
	}

	return false
}

func (h *Handler) Wrap(fn WrapperFunc) {
	e := h.exec
	h.exec = func(w http.ResponseWriter, r *http.Request) {

		// call wrapper logic
		r = fn(r)

		if checkWrapperError(w, r) {
			// error occured and has already responded
			return
		}

		// call core logic
		e(w, r)
	}
}
