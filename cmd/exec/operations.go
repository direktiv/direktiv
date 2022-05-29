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
)

func addAuthHeaders(req *http.Request) {
	req.Header.Add("apikey", apiKey)
	req.Header.Add("Direktiv-Token", authToken)
}

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

func updateRemoteWorkflow(url string, localPath string) error {

	urlUpdate := fmt.Sprintf("%s?op=update-workflow", url)
	urlCreate := fmt.Sprintf("%s?op=create-workflow", url)

	buf, err := safeLoadFile(localPath)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}

	updateFailed := false
	url = urlUpdate

retry:

	req, err := http.NewRequest(
		http.MethodPost,
		urlUpdate,
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
