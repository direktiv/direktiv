package main

import (
	"bytes"
	"io"
	"net/http"
)

func errResponse(w http.ResponseWriter, err error) {

	var code int
	var msg string

	switch err {
	default:
		code = http.StatusInternalServerError
		msg = err.Error()
	}

	w.WriteHeader(code)
	io.Copy(w, bytes.NewReader([]byte(msg)))
}
