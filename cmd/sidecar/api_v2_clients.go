package sidecar

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/utils"
)

func getInstanceVariables(ctx context.Context, flowToken string, flowAddr string, ir *functionRequest) (*variablesResponse, int, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables?instanceId=%v", flowAddr, ir.namespace, ir.instanceId)

	return getVariables(ctx, flowToken, addr)
}

func getNamespaceVariables(ctx context.Context, flowToken string, flowAddr string, ir *functionRequest) (*variablesResponse, int, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables", flowAddr, ir.namespace)

	return getVariables(ctx, flowToken, addr)
}

func getWorkflowVariables(ctx context.Context, flowToken string, flowAddr string, ir *functionRequest) (*variablesResponse, int, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables?workflowPath=%v", flowAddr, ir.namespace, ir.workflowPath)

	return getVariables(ctx, flowToken, addr)
}

func getVariables(ctx context.Context, flowToken, addr string) (*variablesResponse, int, error) {
	resp, err := doRequest(ctx, http.MethodGet, flowToken, addr, nil)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	variables := variablesResponse{}

	statusCode, err := handleResponse(resp, func(resp *http.Response) (int, error) {
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&variables); err != nil {
			return resp.StatusCode, fmt.Errorf("failed to decode response body: %w", err)
		}
		return http.StatusOK, nil
	})
	if err != nil {
		return nil, statusCode, err
	}

	return &variables, resp.StatusCode, nil
}

func getReferencedFile(ctx context.Context, flowToken, flowAddr, namespace string, path string) ([]byte, int, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/files/%v", flowAddr, namespace, path)
	var d []byte

	resp, err := doRequest(ctx, http.MethodGet, flowToken, addr, nil)
	if resp.StatusCode == http.StatusNotFound {
		// some very special magic
		return nil, resp.StatusCode, &RessourceNotFoundError{
			Key:   path,
			Scope: "file",
		}
	}
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to do request: %w", err)
	}
	statusCode, err := handleResponse(resp, func(resp *http.Response) (int, error) {
		var respData decodedFilesResponse
		if resp.Body == nil {
			return http.StatusInternalServerError, fmt.Errorf("unexpected failure: response body is nil")
		}
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&respData); err != nil {
			target := &json.UnmarshalTypeError{}
			if errors.As(err, &target) {
				return http.StatusBadRequest, fmt.Errorf("invalid response format: %w", err)
			}
			return http.StatusInternalServerError, fmt.Errorf("failed to decode response: %w", err)
		}

		if len(respData.Data.Data) > 0 {
			d, err = base64.StdEncoding.DecodeString(respData.Data.Data)
			if err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to decode base64 data: %w", err)
			}
		}
		return http.StatusOK, nil
	})
	if err != nil {
		return nil, statusCode, err
	}

	return d, statusCode, nil
}

func getVariableMetaFromFlow(ctx context.Context, flowToken string, flowAddr string, ir *functionRequest, scope, key string) (variable, int, error) {
	var varResp *variablesResponse
	var err error
	var typ string
	statusCode := http.StatusOK

	// Determine scope and retrieve variables
	switch scope {
	case utils.VarScopeInstance:
		varResp, statusCode, err = getInstanceVariables(ctx, flowToken, flowAddr, ir)
		if err != nil {
			return variable{}, statusCode, fmt.Errorf("failed to get instance variables: %w", err)
		}
		typ = "instance-variable"

	case utils.VarScopeWorkflow:
		varResp, statusCode, err = getWorkflowVariables(ctx, flowToken, flowAddr, ir)
		if err != nil {
			return variable{}, statusCode, fmt.Errorf("failed to get workflow variables: %w", err)
		}
		typ = "workflow-variable"

	case utils.VarScopeNamespace:
		varResp, statusCode, err = getNamespaceVariables(ctx, flowToken, flowAddr, ir)
		if err != nil {
			return variable{}, statusCode, fmt.Errorf("failed to get namespace variables: %w", err)
		}
		typ = "namespace-variable"

	default:
		return variable{}, statusCode, fmt.Errorf("unknown scope: %s", scope)
	}

	idx := slices.IndexFunc(varResp.Data, func(e variable) bool { return e.Typ == typ && e.Name == key })
	if idx < 0 {
		return variable{}, statusCode, &RessourceNotFoundError{Key: key, Scope: scope}
	}

	return varResp.Data[idx], statusCode, nil
}

func getVariableDataViaID(ctx context.Context, flowToken string, flowAddr string, namespace string, id string) (variable, error) {
	addr := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables/%v", flowAddr, namespace, id)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return variable{}, err
	}
	req.Header.Set("Direktiv-Token", flowToken)
	resp, err := client.Do(req)
	if err != nil {
		return variable{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return variable{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	v := variable{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&v); err != nil {
		return variable{}, err
	}

	return v, nil
}

func postVarData(ctx context.Context, flowToken string, flowAddr string, namespace string, body createVarRequest) (int, error) {
	reqD, err := json.Marshal(body)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to marshal request body: %w", err)
	}
	read := bytes.NewReader(reqD)
	url := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables", flowAddr, namespace)

	resp, err := doRequest(ctx, http.MethodPost, flowToken, url, read)
	if err != nil {
		return resp.StatusCode, err
	}

	if statusCode, err := handleResponse(resp, nil); err != nil {
		return statusCode, err
	}

	return http.StatusOK, nil
}

func patchVarData(ctx context.Context, flowToken string, flowAddr string, namespace string, id string, body datastore.RuntimeVariablePatch) (int, error) {
	reqD, err := json.Marshal(body)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to marshal request body: %w", err)
	}
	read := bytes.NewReader(reqD)
	url := fmt.Sprintf("http://%v/api/v2/namespaces/%v/variables/%v", flowAddr, namespace, id)

	resp, err := doRequest(ctx, http.MethodPatch, flowToken, url, read)
	if err != nil {
		return resp.StatusCode, err
	}

	if statusCode, err := handleResponse(resp, nil); err != nil {
		return statusCode, err
	}

	return http.StatusOK, nil
}

func doRequest(ctx context.Context, method, flowToken, url string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Set("Direktiv-Token", flowToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

func handleResponse(resp *http.Response, next func(resp *http.Response) (int, error)) (int, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErrorResp apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErrorResp); err != nil {
			if err == io.EOF {
				return resp.StatusCode, fmt.Errorf("empty error response body")
			}
			return http.StatusInternalServerError, fmt.Errorf("failed to decode error response: %w", err)
		}
		return resp.StatusCode, fmt.Errorf("API error: code %v - message: %v", apiErrorResp.Error.Code, apiErrorResp.Error.Message)
	}

	if next != nil {
		return next(resp)
	}

	return http.StatusOK, nil
}
