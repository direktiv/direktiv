package workflows

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Workflow-related commands",
}

var setReadonlyCmd = &cobra.Command{
	Use:   "toggle-lock DIR_PATH",
	Short: "Sets local directory to readonly or writable in Direktiv server.",
	Long: `Sets local directory to readonly or writable in Direktiv server.

EXAMPLE:
toggle-lock . --addr http://192.168.1.1 --namespace admin --writable
toggle-lock path/to/git/subfolder --namespace admin`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		r, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			root.Fail("could not access recursive flag: %v", err)
		}

		writable, err := cmd.Flags().GetBool("writable")
		if err != nil {
			root.Fail("could not access writable flag: %v", err)
		}

		pathsToUpdate, err := getImpactedFiles(args[0], false, r)
		if err != nil {
			root.Fail("could not calculate impacted files: %v", err)
		}

		for i := range pathsToUpdate {
			p := pathsToUpdate[i]

			path := strings.TrimPrefix(root.GetPath(p), "/")
			path = strings.TrimPrefix(path, ".")
			url := fmt.Sprintf("%s/tree/%s", root.UrlPrefix, path)

			err := switchGitStatus(url, writable)
			if err != nil && !errors.Is(err, ErrNotGit) && !errors.Is(err, ErrNotFound) {
				root.Printlog("can not switch git status: %s", err.Error())
			}

		}

	},
}

func getImpactedFiles(start string, filesAllowed, recursive bool) ([]string, error) {

	pathsToUpdate := make([]string, 0)

	pathStat, err := os.Stat(start)
	if err != nil {
		return pathsToUpdate, fmt.Errorf("could not access path: %w", err)
	}

	if filesAllowed || pathStat.IsDir() {
		if recursive {
			err := filepath.Walk(start,
				func(localPath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if filesAllowed && !info.IsDir() && info.Name() != ".direktiv.yaml" {
						if (strings.HasSuffix(localPath, ".yaml") || strings.HasSuffix(localPath, ".yml")) && !(strings.Contains(localPath, ".yaml.") || strings.Contains(localPath, ".yml.")) {
							pathsToUpdate = append(pathsToUpdate, localPath)
						}
					} else if !filesAllowed && info.IsDir() && !strings.Contains(localPath, ".git") {
						pathsToUpdate = append(pathsToUpdate, localPath)
					}

					return nil
				})

			if err != nil {
				return pathsToUpdate, fmt.Errorf("recursive search could not access path: %w", err)
			}
		} else {
			pathsToUpdate = append(pathsToUpdate, start)
		}
	} else {
		if filesAllowed {
			pathsToUpdate = append(pathsToUpdate, start)
		} else {
			return pathsToUpdate, fmt.Errorf("only directories allowed")
		}
	}

	return pathsToUpdate, nil
}

func switchGitStatus(url string, writable bool) error {

	url = strings.TrimSuffix(url, "/")
	isReadOnly, err := getNodeReadOnly(url)
	if err != nil {
		return err
	}

	operation := "unlock-mirror"

	if writable {
		root.Printlog("set writable for %s", url)

		if !isReadOnly {
			return fmt.Errorf("already writable")
		}

		operation = "lock-mirror"
	} else {
		root.Printlog("set readonly for %s", url)
		if isReadOnly {
			return fmt.Errorf("already read-only")
		}
	}

	urlLockMirror := fmt.Sprintf("%s?op=%s", url, operation)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		urlLockMirror,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}

	root.AddAuthHeaders(req)

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

	return nil

}

var ErrNotFound = errors.New("resource was not found")
var ErrNodeIsReadOnly = errors.New("resource is read-only")
var ErrNotGit = errors.New("resource is not a git folder")

type node struct {
	Namespace string `json:"namespace"`
	Node      struct {
		CreatedAt    time.Time     `json:"createdAt"`
		UpdatedAt    time.Time     `json:"updatedAt"`
		Name         string        `json:"name"`
		Path         string        `json:"path"`
		Parent       string        `json:"parent"`
		Type         string        `json:"type"`
		Attributes   []interface{} `json:"attributes"`
		Oid          string        `json:"oid"`
		ReadOnly     bool          `json:"readOnly"`
		ExpandedType string        `json:"expandedType"`
	} `json:"node"`
}

// getNodeReadOnly : Returns if node at path is read only.
func getNodeReadOnly(url string) (bool, error) {

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("failed to create request file: %w", err)
	}

	root.AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return false, fmt.Errorf("failed to get node information, request was unauthorized")
		}

		if resp.StatusCode == http.StatusNotFound {
			return false, ErrNotFound
		}

		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return false, fmt.Errorf("failed to get node information, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return false, fmt.Errorf("failed to get node information, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	var n node
	err = json.Unmarshal(data, &n)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// that should really not happen
	if n.Node.ExpandedType != "git" {
		return false, ErrNotGit
	}

	return n.Node.ReadOnly, nil
}

var pushCmd = &cobra.Command{
	Use:   "push WORKFLOW_PATH|DIR_PATH",
	Short: "Pushes local workflow or dir to remote direktiv server. This process will update your latest remote resource to your local WORKFLOW_PATH|DIR_PATH file",
	Long: `"Pushes local workflow or dir to remote direktiv server. This process will update your latest remote resource to your local WORKFLOW_PATH|DIR_PATH file.
Pushing local directory cannot be used with config flag. Config must be found automatically to determine folder structure.

EXAMPLE: push helloworld.yaml --addr http://192.168.1.1 --namespace admin

Variables will also be uploaded if they are prefixed with your local workflow name
EXAMPLE:
  dir: /pwd
        /helloworld.yaml
        /helloworld.yaml.data.json
Executing: push helloworld.yaml --addr http://192.168.1.1 --namespace admin --path helloworld
Will update the helloworld workflow and set the remote workflow variable 'data.json' to the contents of '/helloworld.yaml.data.json'
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		pathsToUpdate, err := getImpactedFiles(args[0], true, true)
		if err != nil {
			root.Fail("could not calculate impacted files: %v", err)
		}

		relativeDir := root.GetConfigPath()

		// cull using ignores
		x := make([]string, 0)
		for i := range pathsToUpdate {

			p := pathsToUpdate[i]
			path := root.GetRelativePath(relativeDir, p)

			shouldAdd := true
			for _, g := range root.Globbers {
				if g.Match(path) {
					root.Printlog("ignoring workflow %s", path)
					shouldAdd = false
					break
				}
			}

			if shouldAdd {
				x = append(x, p)
			}
		}

		root.Printlog("found %v local workflow/s to update\n", len(x))

		for i := range x {
			wf := x[i]
			path := root.GetRelativePath(relativeDir, wf)
			path = root.GetPath(path)

			root.Printlog("pushing workflow %s", path)

			// push local variables
			err = updateLocalVars(wf, path)
			if err != nil {
				fmt.Printf("can not update variables: %s\n", err.Error())
			}

			err := updateRemoteWorkflow(path, wf)
			if err != nil {
				fmt.Printf("can not update workflow: %s\n", err.Error())
			}

		}

	},
}

func updateLocalVars(wf, path string) error {
	// push local variables
	localVars, err := getLocalWorkflowVariables(wf)
	if err != nil {
		return fmt.Errorf("failed to get local variable files: %w\n", err)
	}

	if len(localVars) > 0 {
		root.Printlog("found %v local variables to push to remote\n", len(localVars))

		for i := range localVars {
			v := localVars[i]
			varName := filepath.ToSlash(strings.TrimPrefix(v, wf+"."))
			root.Printlog("      updating remote workflow variable: '%s'\n", varName)
			err = setRemoteWorkflowVariable(path, varName, v)
			if err != nil {
				return fmt.Errorf("failed to set remote variable file: %w\n", err)
			}
		}
	}

	return nil
}

func setRemoteWorkflowVariable(wf string, varName string, varPath string) error {

	varData, err := root.SafeLoadFile(varPath)
	if err != nil {
		return fmt.Errorf("failed to load variable file: %w", err)
	}

	urlWorkflow := fmt.Sprintf("%s/tree/%s", root.UrlPrefix, strings.TrimPrefix(wf, "/"))

	url := urlWorkflow + "?op=set-var&var=" + varName

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		url,
		varData,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	root.AddAuthHeaders(req)

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

func updateRemoteWorkflow(path string, localPath string) error {

	root.Printlog("updating namespace: '%s' workflow: '%s'\n", root.GetNamespace(), path)

	isReadOnly, err := getClosestNodeReadOnly(path)
	if err != nil && !errors.Is(err, ErrNotFound) {
		log.Fatalf("Failed to get node : %v", err)
	}

	if isReadOnly {
		return ErrNodeIsReadOnly
	}

	err = recurseMkdirParent(path)
	if err != nil {
		return fmt.Errorf("Failed to create parent directory: %w", err)
	}

	urlWorkflow := fmt.Sprintf("%s/tree/%s", root.UrlPrefix, strings.TrimPrefix(path, "/"))

	urlUpdate := fmt.Sprintf("%s?op=update-workflow", urlWorkflow)
	urlCreate := fmt.Sprintf("%s?op=create-workflow", urlWorkflow)

	buf, err := root.SafeLoadFile(localPath)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}

	data, err := io.ReadAll(buf)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}

	doRequest := func(updateURL, methodIn string, dataIn []byte) (int, string, error) {

		req, err := http.NewRequestWithContext(
			context.Background(),
			methodIn,
			updateURL,
			bytes.NewReader(dataIn),
		)

		if err != nil {
			return 0, "", fmt.Errorf("failed to create request file: %w", err)
		}

		root.AddAuthHeaders(req)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, "", fmt.Errorf("failed to send request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, "",
				fmt.Errorf("failed to update workflow, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(respBody))
		}

		return resp.StatusCode, string(respBody), nil

	}

	// code, err := doRequest(urlUpdate, http.MethodPut)
	// if code == http.StatusNotFound {
	code, body, err := doRequest(urlCreate, http.MethodPut, data)
	if code == http.StatusConflict {
		code, body, err = doRequest(urlUpdate, http.MethodPost, data)
	}

	if err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	if code != http.StatusOK {
		if code == http.StatusUnauthorized {
			return fmt.Errorf("failed to update workflow, request was unauthorized")
		}

		return fmt.Errorf("failed to update workflow, server responded with %d\n------DUMPING ERROR BODY ------\n%s", code, body)
	}

	return nil
}

func recurseMkdirParent(path string) error {

	dirs := strings.Split(filepath.Dir(path), "/")

	for i := range dirs {
		createPath := strings.Join(dirs[:i+1], "/")
		urlDir := fmt.Sprintf("%s/tree/%s?op=create-directory", root.UrlPrefix, strings.Trim(createPath, "/"))
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPut,
			urlDir,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request file: %w", err)
		}

		root.AddAuthHeaders(req)

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

	}

	return nil

}

// getClosestNodeReadOnly : Recursively searches upwards to find closest
// existing node and returns whether it is read only.
func getClosestNodeReadOnly(path string) (bool, error) {

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	dirs := strings.Split(filepath.Dir(path), "/")

	for i := range dirs {
		check := strings.Join(dirs[:(len(dirs)-i)], "/")
		urlDir := fmt.Sprintf("%s/tree/%s", root.UrlPrefix, strings.Trim(check, "/"))
		isReadOnly, err := getNodeReadOnly(urlDir)
		if err != nil && !errors.Is(err, ErrNotGit) {
			return true, err
		}
		if isReadOnly {
			return true, nil
		}
	}

	return false, nil
}

var (
	outputFlag     string
	execNoPushFlag bool
	input          string
)

type logResponse struct {
	PageInfo struct {
		TotalCount int `json:"total"`
		Limit      int `json:"limit"`
		Offset     int `json:"offset"`
	} `json:"pageInfo"`
	Results []struct {
		T   time.Time `json:"t"`
		Msg string    `json:"msg"`
	} `json:"results"`
	Namespace string `json:"namespace"`
	Instance  string `json:"instance"`
}

type instanceResponse struct {
	Namespace string `json:"namespace"`
	Instance  struct {
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
		ID           string    `json:"id"`
		As           string    `json:"as"`
		Status       string    `json:"status"`
		ErrorCode    string    `json:"errorCode"`
		ErrorMessage string    `json:"errorMessage"`
	} `json:"instance"`
	InvokedBy string   `json:"invokedBy"`
	Flow      []string `json:"flow"`
	Workflow  struct {
		Path     string `json:"path"`
		Name     string `json:"name"`
		Parent   string `json:"parent"`
		Revision string `json:"revision"`
	} `json:"workflow"`
}

var execCmd = &cobra.Command{
	Use:   "exec WORKFLOW_PATH",
	Short: "Remotely execute direktiv workflows with local files. This process will update your latest remote workflow to your local WORKFLOW_PATH file",
	Long: `Remotely execute direktiv workflows with local files. This process will update your latest remote workflow to your local WORKFLOW_PATH file.

EXAMPLE: exec helloworld.yaml --addr http://192.168.1.1 --namespace admin --path helloworld

Variables will also be uploaded if they are prefixed with your local workflow name
EXAMPLE:
  dir: /pwd
        /helloworld.yaml
        /helloworld.yaml.data.json
Executing: exec helloworld.yaml --addr http://192.168.1.1 --namespace admin --path helloworld
Will update the helloworld workflow and set the remote workflow variable 'data.json' to the contents of '/helloworld.yaml.data.json'
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		instanceStatus := "pending"

		relativeDir := root.GetConfigPath()
		path := root.GetRelativePath(relativeDir, args[0])
		path = root.GetPath(path)

		fmt.Printf("PATH %v %v\n", path, args[0])

		if !execNoPushFlag {
			err := updateRemoteWorkflow(path, args[0])
			if err != nil {
				fmt.Printf("can not execute workflow: %v\n", err)
			}
		} else {
			root.Printlog("skipping updating namespace: '%s' workflow: '%s'\n", root.GetNamespace(), path)
		}

		err := updateLocalVars(args[0], path)
		if err != nil {
			fmt.Printf("can not update variables: %s\n", err.Error())
		}

		urlExecute := fmt.Sprintf("%s/tree/%s?op=execute&ref=latest", root.UrlPrefix, strings.TrimPrefix(path, "/"))
		instanceDetails, err := executeWorkflow(urlExecute)
		if err != nil {
			log.Fatalf("Failed to execute workflow: %v\n", err)
		}

		root.Printlog("Successfully Executed Instance: %s\n", instanceDetails.Instance)
		root.Printlog("-------INSTANCE LOGS-------")
		urlLogs := fmt.Sprintf("%s/instances/%s/logs", root.UrlPrefix, instanceDetails.Instance)
		clientLogs := sse.NewClient(urlLogs)
		clientLogs.Connection.Transport = &http.Transport{
			TLSClientConfig: root.GetTLSConfig(),
		}

		root.AddSSEAuthHeaders(clientLogs)

		logsChannel := make(chan *sse.Event)
		err = clientLogs.SubscribeChan("messages", logsChannel)
		if err != nil {
			log.Fatalf("Failed to subscribe to messages channel: %v\n", err)
		}

		// Get Logs
		go func() {
			for {
				msg := <-logsChannel
				if msg == nil {
					break
				}

				// Skip heartbeat
				if len(msg.Data) == 0 {
					continue
				}

				var logResp logResponse
				err = json.Unmarshal(msg.Data, &logResp)
				if err != nil {
					log.Fatalln(err)
				}

				if len(logResp.Results) > 0 {
					for _, edge := range logResp.Results {
						cmd.PrintErrf("%v: %s\n", edge.T.In(time.Local).Format("02 Jan 06 15:04 MST"), edge.Msg)
					}
				}
			}
		}()

		urlInstance := fmt.Sprintf("%s/instances/%s", root.UrlPrefix, instanceDetails.Instance)
		clientInstance := sse.NewClient(urlInstance)
		clientInstance.Connection.Transport = &http.Transport{
			TLSClientConfig: root.GetTLSConfig(),
		}

		root.AddSSEAuthHeaders(clientInstance)

		channelInstance := make(chan *sse.Event)
		err = clientInstance.SubscribeChan("messages", channelInstance)
		if err != nil {
			root.Fail("Failed to subscribe to messages channel: %v", err)
		}

		for {
			msg := <-channelInstance
			if msg == nil {
				break
			}

			// Skip heartbeat
			if len(msg.Data) == 0 {
				continue
			}

			var instanceResp instanceResponse
			err = json.Unmarshal(msg.Data, &instanceResp)
			if err != nil {
				log.Fatalf("Failed to read instance response: %v\n", err)
			}

			if instanceResp.Instance.Status != instanceStatus {
				time.Sleep(500 * time.Millisecond)
				instanceStatus = instanceResp.Instance.Status
				clientLogs.Unsubscribe(logsChannel)
				clientInstance.Unsubscribe(channelInstance)
				break

			}
		}

		root.Printlog("instance completed with status: %s\n", instanceStatus)
		urlOutput := fmt.Sprintf("%s/instances/%s/output", root.UrlPrefix, instanceDetails.Instance)

		output, err := getOutput(urlOutput)
		if outputFlag != "" {
			err = os.WriteFile(outputFlag, output, 0600)
			if err != nil {
				log.Fatalf("Failed to write output file: %v\n", err)
			}
		} else {
			cmd.PrintErrln("------INSTANCE OUTPUT------")
			fmt.Println(string(output))
		}
	},
}

type executeResponse struct {
	Instance string `json:"instance,omitempty"`
}

func executeWorkflow(url string) (executeResponse, error) {
	var instanceDetails executeResponse

	// Read input data from flag file
	inputData, err := root.SafeLoadFile(input)
	if err != nil {
		return instanceDetails, fmt.Errorf("Failed to load input file: %w", err)
	}

	// If inputData is empty attempt to read from stdin
	if inputData.Len() == 0 {
		inputData, err = root.SafeLoadStdIn()
		if err != nil {
			return instanceDetails, fmt.Errorf("Failed to load stdin: %w", err)
		}
	}

	// If no file or stdin input data was provided, set data to {}
	if inputData.Len() == 0 {
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

	root.AddAuthHeaders(req)

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

type instanceOutput struct {
	Namespace string `json:"namespace"`
	Instance  struct {
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
		ID           string    `json:"id"`
		As           string    `json:"as"`
		Status       string    `json:"status"`
		ErrorCode    string    `json:"errorCode"`
		ErrorMessage string    `json:"errorMessage"`
	} `json:"instance"`
	Data string `json:"data"`
}

func getOutput(url string) ([]byte, error) {
	var output instanceOutput

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	root.AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("failed to get instance output, request was unauthorized")
		}

		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return nil, fmt.Errorf("failed to get instance output, server responded with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return nil, fmt.Errorf("failed to get instance output, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)

	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &output)
	if err != nil {
		return nil, err
	}

	outputStr, err := base64.StdEncoding.DecodeString(output.Data)
	return outputStr, err

}

func init() {
	root.RootCmd.AddCommand(workflowCmd)

	workflowCmd.AddCommand(setReadonlyCmd)
	workflowCmd.AddCommand(pushCmd)
	workflowCmd.AddCommand(execCmd)

	setReadonlyCmd.PersistentFlags().Bool("recursive", false, "Recursively set folders read-only.")
	setReadonlyCmd.PersistentFlags().Bool("writable", false, "Sets the folders to writable, readonly otherwise.")

	execCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Path where to write instance output. If unset output will be written to screen")
	execCmd.Flags().StringVarP(&input, "input", "i", "", "Path to file to be used as input data for executed workflow. If unset, stdin will be used as input data if available.")
	execCmd.Flags().BoolVar(&execNoPushFlag, "no-push", false, "If set will skip updating and just execute the workflow.")
}
