package webapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func writeError(w http.ResponseWriter, err *Error) {
	httpStatus := http.StatusInternalServerError

	if strings.HasPrefix(err.Code, "credentials_") {
		httpStatus = http.StatusForbidden
	}
	if strings.HasPrefix(err.Code, "request_") {
		httpStatus = http.StatusBadRequest
	}
	if strings.HasPrefix(err.Code, "request_path_not_found") {
		httpStatus = http.StatusNotFound
	}

	if httpStatus == http.StatusInternalServerError {
		fmt.Printf("internal error: %s\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	payLoad := struct {
		Error *Error
	}{
		Error: err,
	}

	_ = json.NewEncoder(w).Encode(payLoad)
}

func writeFunctionsError(w http.ResponseWriter, err error) {
	//if errors.Is(err, store.ErrBadKey) {
	//	writeError(c, &Error{
	//		Code:    "request_key",
	//		Message: "empty or invalid field 'key'",
	//	})
	//	return
	//}

	writeError(w, &Error{
		Code:    "internal",
		Message: "internal error",
	})
}
