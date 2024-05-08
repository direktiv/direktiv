package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

type Error struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Validation map[string]string `json:"validation,omitempty"`
}

func writeError(w http.ResponseWriter, err *Error) {
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

	httpStatus := http.StatusInternalServerError

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

	slog.Error("internal", "err", err)
}

func writeBadrequestError(w http.ResponseWriter, err error) {
	writeError(w, &Error{
		Code:    "request",
		Message: "bad request",
	})

	slog.Error("internal", "err", err)
}

func writeNotJSONError(w http.ResponseWriter, err error) {
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
	if errors.Is(err, datastore.ErrInvalidRuntimeVariableName) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "field name has invalid string",
		})

		return
	}
	if errors.Is(err, datastore.ErrInvalidNamespaceName) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "invalid namespace name",
		})

		return
	}
	if errors.Is(err, datastore.ErrDuplication) {
		writeError(w, &Error{
			Code:    "resource_already_exists",
			Message: "resource already exists",
		})

		return
	}

	if errors.Is(err, datastore.ErrDuplicatedNamespaceName) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "namespace name already used",
		})

		return
	}

	var vErrs datastore.ValidationError
	if errors.As(err, &vErrs) {
		writeError(w, &Error{
			Code:       "request_data_invalid",
			Message:    "request data has invalid fields",
			Validation: vErrs,
		})

		return
	}

	writeInternalError(w, err)
}

func writeFileStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, filestore.ErrNotFound) {
		writeError(w, &Error{
			Code:    "resource_not_found",
			Message: "filesystem path is not found",
		})

		return
	}
	if errors.Is(err, filestore.ErrPathAlreadyExists) {
		writeError(w, &Error{
			Code:    "resource_already_exists",
			Message: "filesystem path already exists",
		})

		return
	}
	if errors.Is(err, filestore.ErrNoParentDirectory) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "filesystem path has no parent directory",
		})

		return
	}
	if errors.Is(err, filestore.ErrFileTypeIsDirectory) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "filesystem path is a directory",
		})

		return
	}
	if errors.Is(err, filestore.ErrInvalidPathParameter) {
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: "filesystem path is invalid",
		})

		return
	}

	writeInternalError(w, err)
}
