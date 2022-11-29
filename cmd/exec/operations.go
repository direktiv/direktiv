package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/util"
)

var ErrNotFound = errors.New("resource was not found")
var ErrNodeIsReadOnly = errors.New("resource is read-only")

func setRemoteWorkflowVariable(wfURL string, varName string, varPath string) error {
	varData, err := safeLoadFile(varPath)
	if err != nil {
		return fmt.Errorf("failed to load variable file: %w", err)
	}

	url := wfURL + "?op=set-var&var=" + varName

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		url,
		varData,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("failed to set workflow var, request was unauthorized")
		}

		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to set workflow var, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to set workflow var, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	return err
}

func getLocalWorkflowVariables(absPath string) ([]string, error) {
	varFiles := make([]string, 0)
	wfFileName := filepath.Base(absPath)
	dirPath := filepath.Dir(absPath)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return varFiles, fmt.Errorf("failed to read dir: %w", err)
	}

	// Find all var files: {LOCAL_PATH}/{WF_FILE}.{VAR}
	for _, file := range files {
		fName := file.Name()
		if !file.IsDir() && fName != wfFileName && strings.HasPrefix(fName, wfFileName) {
			varFiles = append(varFiles, filepath.Join(dirPath, fName))
		}
	}

	return varFiles, nil
}

func recurseMkdirParent(path string) error {

	dir, _ := filepath.Split(path)
	if dir == "" || dir == "/" {
		return nil
	}

	dir = strings.TrimSuffix(dir, "/")

	err := recurseMkdirParent(dir)
	if err != nil {
		return err
	}

	urlDir := fmt.Sprintf("%s/tree/%s", urlPrefix, strings.Trim(dir, "/"))
	urlMkdir := fmt.Sprintf("%s?op=create-directory", urlDir)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		urlMkdir,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("failed to create parent, request was unauthorized")
		}

		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to create parent, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to create parent, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	return err

}

// getClosestNodeReadOnly : Recursively searches upwards to find closest
// existing node and returns whether it is read only.
func getClosestNodeReadOnly(path string) (bool, string, error) {
	isReadOnly, nodeType, err := getNodeReadOnly(path)

	if errors.Is(err, ErrNotFound) {
		dir, _ := filepath.Split(path)
		dir = strings.TrimSuffix(dir, "/")

		return getClosestNodeReadOnly(dir)
	}

	return isReadOnly, nodeType, err
}

// getNodeReadOnly : Returns if node at path is read only.
func getNodeReadOnly(path string) (bool, string, error) {
	urlGetNode := fmt.Sprintf("%s/tree/%s", urlPrefix, strings.TrimPrefix(path, "/"))

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		urlGetNode,
		nil,
	)
	if err != nil {
		return false, "", fmt.Errorf("failed to create request file: %w", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return false, "", fmt.Errorf("failed to get node information, request was unauthorized")
		}

		if resp.StatusCode == http.StatusNotFound {
			return false, "", ErrNotFound
		}

		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return false, "", fmt.Errorf("failed to get node information, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return false, "", fmt.Errorf("failed to get node information, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("failed to read response: %w", err)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	if err != nil {
		return false, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	x, exists := m["node"]
	if !exists {
		return false, "", fmt.Errorf("unexpected response: %v", string(data))
	}

	m2, ok := x.(map[string]interface{})
	if !ok {
		return false, "", fmt.Errorf("unexpected response: %v", string(data))
	}

	x, exists = m2["readOnly"]
	if !exists {
		return false, "", fmt.Errorf("unexpected response: %v", string(data))
	}

	ro, ok := x.(bool)
	if !ok {
		return false, "", fmt.Errorf("unexpected response: %v", string(data))
	}

	x, exists = m2["expandedType"]
	if !exists {
		return ro, "", fmt.Errorf("unexpected response: %v", string(data))
	}

	et, ok := x.(string)
	if !ok {
		return ro, "", fmt.Errorf("unexpected response: %v", string(data))
	}

	return ro, et, nil
}

func setWritable(path string, value bool) error {

	dir, _ := filepath.Split(path)
	dir = strings.TrimSuffix(dir, "/")

	urlWorkflow = fmt.Sprintf("%s/tree/%s", urlPrefix, strings.TrimPrefix(path, "/"))

	isReadOnly, nodeType, err := getNodeReadOnly(path)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			err = setWritable(dir, value)
			if err != nil {
				return err
			}
			return nil
		}

		return err
	}

	if isReadOnly != value {
		return nil
	}

	switch nodeType {
	case util.InodeTypeGit:

		var operation string
		if value {
			operation = "lock-mirror"
		} else {
			operation = "unlock-mirror"
		}

		urlLockMirror := fmt.Sprintf("%s?op=%s", urlWorkflow, operation)

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			urlLockMirror,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request file: %w", err)
		}

		addAuthHeaders(req)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusUnauthorized {
				return errors.New("failed to modify node, request was unauthorized")
			}

			errBody, err := io.ReadAll(resp.Body)
			if err == nil {
				return fmt.Errorf("failed to modify node, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
			}

			return fmt.Errorf("failed to modify node, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
		}

	default:

		err = setWritable(dir, value)
		if err != nil {
			return err
		}

	}

	return nil

}

func updateRemoteWorkflow(path string, localPath string) error {

	printlog("updating namespace: '%s' workflow: '%s'\n", getNamespace(), path)

	isReadOnly, _, err := getClosestNodeReadOnly(path)
	if err != nil && !errors.Is(err, ErrNotFound) {
		log.Fatalf("Failed to get node : %v", err)
	}

	if isReadOnly {
		return ErrNodeIsReadOnly
	}

	err = recurseMkdirParent(path)
	if err != nil {
		log.Fatalf("Failed to create parent directory: %v", err)
	}

	urlWorkflow = fmt.Sprintf("%s/tree/%s", urlPrefix, strings.TrimPrefix(path, "/"))

	urlUpdate := fmt.Sprintf("%s?op=update-workflow", urlWorkflow)
	urlCreate := fmt.Sprintf("%s?op=create-workflow", urlWorkflow)

	buf, err := safeLoadFile(localPath)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}
	data, err := io.ReadAll(buf)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}

	updateFailed := false
	url := urlUpdate
	method := http.MethodPost

retry:

	req, err := http.NewRequestWithContext(
		context.Background(),
		method,
		url,
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("failed to update workflow, request was unauthorized")
		}

		if resp.StatusCode == http.StatusNotFound && !updateFailed {
			updateFailed = true
			url = urlCreate
			method = http.MethodPut
			goto retry
		}
		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to update workflow, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to update workflow, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	return err
}

func executeWorkflow(url string) (executeResponse, error) {
	var instanceDetails executeResponse

	// Read input data from flag file
	inputData, err := safeLoadFile(input)
	if err != nil {
		log.Fatalf("Failed to load input file: %v", err)
	}

	// If inputData is empty attempt to read from stdin
	if inputData.Len() == 0 {
		inputData, err = safeLoadStdIn()
		if err != nil {
			log.Fatalf("Failed to load stdin: %v", err)
		}
	}

	// If no file or stdin input data was provided, set data to {}
	if inputData.Len() == 0 && inputType == "application/json" {
		inputData = bytes.NewBufferString("{}")
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url,
		inputData,
	)
	if err != nil {
		return instanceDetails, err
	}

	req.Header.Add("Content-Type", inputType)
	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return instanceDetails, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return instanceDetails, fmt.Errorf("failed to execute workflow, request was unauthorized")
		}

		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return instanceDetails, fmt.Errorf("failed to execute workflow, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return instanceDetails, fmt.Errorf("failed to execute workflow, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return instanceDetails, err
	}

	err = json.Unmarshal(body, &instanceDetails)
	return instanceDetails, err

}

func executeEvent(url string) (string, error) {

	// Read input data from flag file
	inputDataOsFile, err := os.Open(localAbsPath)

	if err != nil {
		return "", err
	}

	defer inputDataOsFile.Close()
	byteResult, err := io.ReadAll(inputDataOsFile)

	if err != nil {
		return "", err
	}

	var event map[string]interface{}
	err = json.Unmarshal(byteResult, &event)

	if err != nil {
		return "", err
	}

	//fill or overwrite inputData if necessary
	if Id != "" {
		event["id"] = Id
	}
	if Source != "" {
		event["source"] = Source
	}
	if Type != "" {
		event["type"] = Type
	}
	if Specversion != "" {
		event["specversion"] = Specversion
	}

	if len(event) == 0 {
		err = errors.New("empty file ")
		return "", err
	}

	eventBody, err := json.Marshal(event)

	if err != nil {
		return "", err
	}

	body := bytes.NewReader(eventBody)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url,
		body,
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", inputType)
	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	//if event already exist just replay the event
	if resp.StatusCode != http.StatusOK {
		eventId := event["id"]
		url := fmt.Sprintf("%s/events/%s/replay", urlPrefix, eventId)

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			url,
			nil,
		)

		if err != nil {
			return "", err
		}

		addAuthHeaders(req)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("failed to send event (rejected by server)")
			return "", err
		}
	}

	x, exist := event["id"]
	if !exist {
		err = errors.New("event id undefined")
		return "", err
	}

	id, ok := x.(string)
	if !ok {
		err = errors.New("event id is not a string")
		return "", err
	}

	return id, err

}

func pingNamespace() error {
	urlGetNode := fmt.Sprintf("%s/tree/", urlPrefix)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		urlGetNode,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping namespace: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("failed to ping namespace, request was unauthorized")
		}

		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to ping namespace, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to ping namespace, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)

	}

	return nil
}
