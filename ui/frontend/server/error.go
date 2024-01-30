package server

import (
	"encoding/json"
	"net/http"
)

type ErrResponse struct {
	Message string `json:"message"`        // user-level status message
	Code    string `json:"code,omitempty"` // error code
	Error   string `json:"error"`          // technical error string
}

const (
	ErrorServerInternal = "server.internal.error"
)

func SendError(w http.ResponseWriter, err error, statusCode int,
	code, message string,
) {
	w.WriteHeader(statusCode)

	er := &ErrResponse{
		Error:   err.Error(),
		Code:    code,
		Message: message,
	}

	b, _ := json.MarshalIndent(er, "", "  ")

	w.Write(b)
}
