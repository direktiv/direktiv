package sidecar

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupAPIForFlow(userServiceURL string, maxResponseSize int, actionCtl *sync.Map) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	// Router for handling external requests.
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		executeFunction(r, w, userServiceURL, maxResponseSize, actionCtl, ActionBuilder{})
	})
	return router
}

func executeFunction(r *http.Request, w http.ResponseWriter, userServiceURL string, maxResponseSize int, actionCtl *sync.Map, actionDeserialize ActionDeserialize) {
	// 1. Validate/Extract Inputs.
	actionID, carrier, err := actionDeserialize.Extract(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 2. Build actionCtl.
	ctx, cancel := context.WithTimeout(r.Context(), carrier.Deadline)
	ctl := actionController{
		cancel:         cancel,
		RequestCarrier: carrier,
	}
	actionCtl.Store(actionID, ctl)

	// 3. Provision.
	filesLocation := filepath.Join(SharedDir, actionID)
	err = writeFiles(filesLocation, ctl.Files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Prepare Request.
	req, err := prepareRequestToUserContainer(ctx, actionID, userServiceURL, filesLocation, ctl.UserInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 5. Execute Request.
	resp, err := executeRequestToUserContainer(maxResponseSize, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	// 6. Response Forwarding.
	responseForwardingToClient(resp).ServeHTTP(w, r)
}

func prepareRequestToUserContainer(ctx context.Context, actionID, userServiceURL, filesLocation string, userInput []byte) (*http.Request, error) {
	// 1. Construct the base HTTP request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, userServiceURL+"?action_id="+actionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	// 2. Add FilesLocationHeader.
	req.Header.Add(FilesLocationHeader, filesLocation)
	// 3. Prepare request body.
	buffer := new(bytes.Buffer)
	_, err = buffer.Write(userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request body: %w", err)
	}
	req.Body = io.NopCloser(buffer)

	return req, nil
}

func executeRequestToUserContainer(maxResponseSize int, req *http.Request) (*http.Response, error) {
	// 1. Execute the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// 2. Basic Response Validation
	if resp.StatusCode >= 400 { // Handle error status codes.
		return nil, handleErrorResponse(resp)
	}

	// 3. Check Response Size (if needed)
	if resp.ContentLength > int64(maxResponseSize) {
		return nil, fmt.Errorf("response exceeds maximum allowed size")
	}

	return resp, nil
}

func responseForwardingToClient(resp *http.Response) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the response from the context

		// Handle non-success status codes (might be done earlier).
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			rC := ResponseCarrier{
				ErrCode: "container_failure",
				Err:     fmt.Errorf("container failed with status %v", resp.StatusCode),
			}
			writeJSON(w, rC)
			return
		}

		// Forward headers from the remote response.
		for header, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(header, value)
			}
		}

		// Copy status code.
		w.WriteHeader(resp.StatusCode)

		// Forward response body.
		_, err := io.Copy(w, resp.Body)
		if err != nil {
			slog.Error("coping resp body", "error", err)
		}
	})
}

// Helper function for error handling.
func handleErrorResponse(resp *http.Response) error {
	errCode := resp.Header.Get(ErrorCodeHeader)
	errMsg := resp.Header.Get(ErrorMessageHeader)

	if errCode != "" {
		return fmt.Errorf("remote service error: %s - %s", errCode, errMsg)
	}

	// Fallback for generic errors if no specific headers are present
	return fmt.Errorf("remote service error: status code %d", resp.StatusCode)
}

// redundant? just flow can just cancel the context
// func cancelAction(r *http.Request, dataMap *sync.Map, config Config) (int, error) {
// 	actionID := r.URL.Query().Get("action_id")
// 	value, loaded := dataMap.Load(actionID)
// 	if !loaded {
// 		return http.StatusInternalServerError, fmt.Errorf("Error action with this ID is not known")
// 	}
// 	action, ok := value.(Action)
// 	if !ok {
// 		return http.StatusInternalServerError, fmt.Errorf("Error Sidecar in invalid state")
// 	}
// 	defer action.cancel()
// 	resp, err := cancelRequest(r.Context(), config.UserServiceURL, actionID)
// 	if err != nil {
// 		return http.StatusInternalServerError, fmt.Errorf("Error Sidecar in invalid state")
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != http.StatusOK {
// 		return http.StatusInternalServerError, fmt.Errorf("Error forwarding request or non-200 status received")
// 	}
// 	return http.StatusOK, nil
// }

// func cancelRequest(ctx context.Context, userServiceURL string, actionID string) (*http.Response, error) {
// 	req, err := http.NewRequest(http.MethodDelete, userServiceURL, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req = req.WithContext(ctx)
// 	req.Header.Add(ActionIDHeader, actionID)
// 	client := &http.Client{}
// 	return client.Do(req)
// }

// func actionHandler(r *http.Request, w http.ResponseWriter, config Config, dataMap *sync.Map) {
// 	ctx, err := contextCreation(r, config.UserServiceURL)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 	}
// 	r.WithContext(ctx)
// 	action, ok := ctx.Value(actionCtxKey).(Action)
// 	if !ok {
// 		http.Error(w, "action,", http.StatusInternalServerError)
// 	}
// 	actionID, ok := ctx.Value(actionIDCtxKey).(string)
// 	if !ok {
// 		http.Error(w, "actionID, ", http.StatusInternalServerError)
// 	}
// 	dataMap.Store(actionID, action)

// 	err = filePreparation(ctx)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 	}
// 	req, err := remoteRequestCreation(ctx)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 	}
// 	err = requestBodyPreparation(ctx, req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 	}
// 	resp, err := requestExecution(ctx, req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 	}
// 	ResponseForwarding(resp).ServeHTTP(w, r)
// 	//
// }
