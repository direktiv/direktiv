package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/grpc/status"
)

func errResponse(w http.ResponseWriter, err error) {

	var code int
	var msg string

	s := status.Convert(err)

	msg = fmt.Sprintf("%s Error: %s", s.Code().String(), s.Message())

	switch s.Code() {
	default:
		code = http.StatusInternalServerError
	}

	w.WriteHeader(code)
	io.Copy(w, bytes.NewReader([]byte(msg)))
}
