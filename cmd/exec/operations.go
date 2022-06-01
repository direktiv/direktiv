package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/util"
)

func setRemoteWorkflowVariable(wfURL string, varName string, varPath string) error {
	varData, err := safeLoadFile(varPath)
	if err != nil {
		return fmt.Errorf("failed to load variable file: %v", err)
	}

	url := wfURL + "?op=set-var&var=" + varName

	req, err := http.NewRequest(
		http.MethodPut,
		url,
		varData,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != 200 {
		errBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to set workflow var, server responsed with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to set workflow var, server responsed with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	return err
}

func getLocalWorkflowVariables(absPath string) ([]string, error) {
	varFiles := make([]string, 0)
	wfFileName := filepath.Base(absPath)
	dirPath := filepath.Dir(absPath)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return varFiles, fmt.Errorf("failed to read dir: %v", err)
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

	req, err := http.NewRequest(
		http.MethodPut,
		urlMkdir,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %v", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != http.StatusConflict {
		errBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to create parent, server responsed with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to create parent, server responsed with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	return err

}

func setWritable(path string) error {

	dir, _ := filepath.Split(path)
	dir = strings.TrimSuffix(dir, "/")

	urlWorkflow = fmt.Sprintf("%s/tree/%s", urlPrefix, strings.TrimPrefix(path, "/"))
	urlGetNode := fmt.Sprintf("%s", urlWorkflow)

	req, err := http.NewRequest(
		http.MethodGet,
		urlGetNode,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %v", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == http.StatusNotFound {
			err = setWritable(dir)
			if err != nil {
				return err
			}
			return nil
		}

		errBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to get node information, server responsed with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to get node information, server responsed with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	x, exists := m["node"]
	if !exists {
		return fmt.Errorf("unexpected response: %v", string(data))
	}

	m2, ok := x.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response: %v", string(data))
	}

	x, exists = m2["readOnly"]
	if !exists {
		return fmt.Errorf("unexpected response: %v", string(data))
	}

	ro, ok := x.(bool)
	if !ok {
		return fmt.Errorf("unexpected response: %v", string(data))
	}

	if ro == false {
		return nil
	}

	x, exists = m2["expandedType"]
	if !exists {
		return fmt.Errorf("unexpected response: %v", string(data))
	}

	et, ok := x.(string)
	if !ok {
		return fmt.Errorf("unexpected response: %v", string(data))
	}

	switch et {
	case util.InodeTypeGit:

		urlLockMirror := fmt.Sprintf("%s?op=lock-mirror", urlWorkflow)

		req, err := http.NewRequest(
			http.MethodPost,
			urlLockMirror,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request file: %v", err)
		}

		addAuthHeaders(req)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %v", err)
		}

		if resp.StatusCode != 200 {
			errBody, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				return fmt.Errorf("failed to get node information, server responsed with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
			}

			return fmt.Errorf("failed to get node information, server responsed with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
		}

	default:

		err = setWritable(dir)
		if err != nil {
			return err
		}

	}

	return nil

}

func updateRemoteWorkflow(path string, localPath string) error {

	printlog("updating namespace: '%s' workflow: '%s'\n", getNamespace(), path)

	err := setWritable(path)
	if err != nil {
		log.Fatalf("Failed to make writable: %v", err)
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
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}

	updateFailed := false
	url := urlUpdate
	method := http.MethodPost

retry:

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %v", err)
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == http.StatusNotFound && !updateFailed {
			updateFailed = true
			url = urlCreate
			method = http.MethodPut
			goto retry
		}
		errBody, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to update workflow, server responsed with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to update workflow, server responsed with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
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

	req, err := http.NewRequest(
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return instanceDetails, err
	}

	err = json.Unmarshal(body, &instanceDetails)
	return instanceDetails, err

}
