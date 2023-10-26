// nolint
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, err *Error) {
	httpStatus := http.StatusInternalServerError

	// access_token_denied
	// access_token_missing
	// access_token_invalid

	// request_path_not_found
	// request_method_not_allowed
	// request_body_not_json
	// resource_not_found
	// resource_already_exists
	// resource_id_invalid

	// request_data_invalid

	httpStatus = http.StatusInternalServerError

	if strings.HasPrefix(err.Code, "access") {
		httpStatus = http.StatusForbidden
	}
	if strings.HasPrefix(err.Code, "request") {
		httpStatus = http.StatusBadRequest
	}
	if strings.HasPrefix(err.Code, "resource") {
		httpStatus = http.StatusBadRequest
	}
	if strings.Contains(err.Code, "not_found") {
		httpStatus = http.StatusNotFound
	}
	if strings.Contains(err.Code, "method_not_allowed") {
		httpStatus = http.StatusMethodNotAllowed
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	payLoad := struct {
		Error *Error `json:"error"`
	}{
		Error: err,
	}

	_ = json.NewEncoder(w).Encode(payLoad)
}

func writeInternalError(w http.ResponseWriter, err error) {
	writeError(w, &Error{
		Code:    "internal",
		Message: "internal server error",
	})

	slog.Error("error", "err", err)
}

func writeNotJsonError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "cannot unmarshal") {
		writeError(w, &Error{
			Code:    "request_body_bad_json_schema",
			Message: "request payload has bad json schema",
		})
		return
	}

	writeError(w, &Error{
		Code:    "request_body_not_json",
		Message: "couldn't parse request payload in json format",
	})
}

func writeDataStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, datastore.ErrNotFound) {
		writeError(w, &Error{
			Code:    "resource_not_found",
			Message: "requested resource is not found",
		})

		return
	}

	writeInternalError(w, err)
}
