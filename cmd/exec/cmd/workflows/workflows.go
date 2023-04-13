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
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Workflow-related commands",
}

func getFiles(start string) ([]string, []string, error) {
	files := make([]string, 0)
	directories := make([]string, 0)
	pathStat, err := os.Stat(start)
	if err != nil {
		return make([]string, 0), make([]string, 0), fmt.Errorf("could not access path: %w", err)
	}

	if !pathStat.IsDir() {
		return make([]string, 0), make([]string, 0), fmt.Errorf("only directories allowed")
	}

	err = filepath.Walk(start,
		func(localPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			absPath, err := filepath.Abs(localPath)
			if err != nil {
				return err
			}
			if info.IsDir() {
				directories = append(directories, absPath)
			}
			if !info.IsDir() {
				files = append(files, absPath)
			}
			return nil
		})
	if err != nil {
		return make([]string, 0), make([]string, 0), err
	}
	return directories, files, nil
}

func fileHasType(file string, types ...string) bool {
	filter := func(s, suf string) bool {
		return strings.HasSuffix(s, suf)
	}
	found := false
	for _, v := range types {
		found = found || filter(file, v)
	}
	return found
}

func filter(in []string, filter func(string) bool) (out []string) {
	for _, s := range in {
		if filter(s) {
			out = append(out, s)
		}
	}
	return out
}

var (
	ErrNotFound       = errors.New("resource was not found")
	ErrNodeIsReadOnly = errors.New("resource is read-only")
	ErrNotGit         = errors.New("resource is not a git folder")
)

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

var pushCmd = &cobra.Command{
	Use:   "push PATH ...FLAG",
	Short: "Pushes or updates your a local direktiv-project to the server",
	Long: `Push or update your a local direktiv-project to the server. 

PATH MUST point to the root-folder in a direktiv package-format.
package-format example:

	PATH format: 
		PATH/.direktiv.yaml
		PATH/helloworld.yaml
		PATH/helloworld.yaml.data.json
		PATH/more/otherwf.yaml

The configuration MUST be present and located in the direktiv-package root-folder.
!!! The configuration-values in .direktiv.yaml from the CAN be overridden by using the global flags`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dirs, files, err := getFiles(args[0])
		if err != nil {
			root.Fail("could not calculate impacted files: %v", err)
		}

		workflows := make([]string, 0)
		wfVariables := make([]string, 0)
		direktivYAML := ""
		for _, f := range files {
			if fileHasType(f, ".yaml", ".yml") &&
				!fileHasType(f, ".direktiv.yaml") {
				workflows = append(workflows, f)
			}
			if fileHasType(f, ".direktiv.yaml") {
				direktivYAML = f
			}

			isVariable := strings.Contains(f, ".yaml.")
			isVariable = isVariable || strings.Contains(f, ".yml.")
			if isVariable {
				wfVariables = append(wfVariables, f)
			}
			for _, g := range root.Globbers {
				files = filter(files, g.Match)
			}
		}
		executeWf := ""
		if executeFlag != "" {
			for _, v := range files {
				if strings.HasSuffix(v, executeFlag) {
					executeWf = v
				}
			}
		}
		if executeFlag != "" && executeWf == "" {
			root.Fail("workflow %s in exec is not in the project", executeFlag)
		}

		projectRoot := dirs[0]
		dirs = dirs[1:]
		if direktivYAML == "" ||
			projectRoot+"/.direktiv.yaml" != direktivYAML &&
				projectRoot+"/.direktiv.yml" != direktivYAML {
			root.Fail(projectRoot + "/.direktiv.yaml")
		}

		for _, d := range dirs {
			relativePath, err := filepath.Rel(projectRoot, d)
			if err != nil {
				root.Fail("error: %v", err)
			}
			err = createWfDirectory(relativePath)
			if err != nil {
				root.Fail("could not create directory %d err: %v", d, err)
			}
		}

		for _, wf := range workflows {

			root.Printlog("pushing workflow %s", wf)

			err = updateRemoteWorkflow(projectRoot, wf)
			if err != nil {
				fmt.Printf("can not update workflow: %s\n", err.Error())
			}
		}
		for _, v := range wfVariables {
			err := setRemoteWorkflowVariable(projectRoot, v)
			if err != nil {
				root.Fail("failed to set remote variable file: %w\n", err)
			}
		}
		if executeWf == "" {
			return
		}
		execWf, err := filepath.Rel(projectRoot, executeWf)
		if err != nil {
			root.Fail("%w\n", err)
		}
		urlExecute := fmt.Sprintf("%s/tree/%s?op=execute&ref=latest", root.UrlPrefix, strings.TrimPrefix(execWf, "/"))
		resp, err := executeWorkflow(urlExecute)
		if err != nil {
			root.Fail("executing failed %v", err)
		}
		root.Printlog("Successfully Executed Instance: %s\n", resp.Instance)
		urlOutput := root.GetLogs(cmd, resp.Instance, "")
		output, err := getOutput(urlOutput)
		if err != nil {
			fmt.Println(err)
			return
		}
		cmd.PrintErrln("------INSTANCE OUTPUT------")
		fmt.Println(string(output))
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
			root.Printlog("updating remote workflow variable: '%s'\n", filepath.Base(varName))
			err = setRemoteWorkflowVariable(path, v)
			if err != nil {
				return fmt.Errorf("failed to set remote variable file: %w\n", err)
			}
		}
	}

	return nil
}

func setRemoteWorkflowVariable(projectRoot, varPath string) error {
	varData, err := root.SafeLoadFile(varPath)
	if err != nil {
		return fmt.Errorf("failed to load variable file: %w", err)
	}
	v, err := filepath.Rel(projectRoot, varPath)
	if err != nil {
		root.Fail("error: %v", err)
	}
	wf := strings.Split(v, ".yml.")[0] + ".yml"
	if strings.Contains(v, ".yaml.") {
		wf = strings.Split(v, ".yaml.")[0] + ".yaml"
	}
	urlWorkflow := fmt.Sprintf("%s/tree/%s", root.UrlPrefix, strings.TrimPrefix(wf, "/"))
	va := filepath.Base(varPath)
	wfName := filepath.Base(wf)
	varName := strings.Replace(va, wfName+".", "", -1)
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

func updateRemoteWorkflow(projectRoot, workflowFile string) error {
	workflow, err := filepath.Rel(projectRoot, workflowFile)
	if err != nil {
		root.Fail("error: %v", err)
	}
	root.Printlog("updating namespace: '%s' workflow: '%s'\n", root.GetNamespace(), workflow)

	urlWorkflow := fmt.Sprintf("%s/tree/%s", root.UrlPrefix, strings.TrimPrefix(workflow, "/"))

	urlUpdate := fmt.Sprintf("%s?op=update-workflow", urlWorkflow)
	urlCreate := fmt.Sprintf("%s?op=create-workflow", urlWorkflow)

	buf, err := root.SafeLoadFile(workflowFile)
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

func createWfDirectory(dir string) error {
	urlDir := fmt.Sprintf("%s/tree/%s?op=create-directory", root.UrlPrefix, dir)
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
	return nil
}

var (
	executeFlag string
	input       string
)

var execCmd = &cobra.Command{
	Use:   "exec WORKFLOW_PATH",
	Short: "Receives a workflow yaml from stdin and executes it with the given path on the Server.",
	Long: `Receives a workflow yaml from stdin and executes it with the given path on the Server.
	
	EXAMPLE cat workflows/start.yaml | ./direktivctl --addr http://192.168.122.232/ -n ns workflows exec folderA/start
	
	will upload the input from stdin to the locatation on the server folderA as workflow start and execute it.
	If you need to redirect the output to a file use: 
	
	cat workflows/start.yaml | ./direktivctl --addr http://192.168.122.232/ -n ns workflows exec folderA/start > mylog.log`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		urlExecute := fmt.Sprintf("%s/tree/%s?op=execute&ref=latest", root.UrlPrefix, strings.TrimPrefix(args[0], "/"))

		instanceDetails, err := executeWorkflow(urlExecute)
		if err != nil {
			log.Fatalf("Failed to execute workflow: %v\n", err)
		}
		root.Printlog("Successfully Executed Instance: %s\n", instanceDetails.Instance)
		urlOutput := root.GetLogs(cmd, instanceDetails.Instance, "")
		output, err := getOutput(urlOutput)
		if err != nil {
			fmt.Println(err)
			return
		}
		cmd.PrintErrln("------INSTANCE OUTPUT------")
		fmt.Println(string(output))
	},
}

type executeResponse struct {
	Instance string `json:"instance,omitempty"`
}

func executeWorkflow(url string) (executeResponse, error) {
	var instanceDetails executeResponse
	var inputData *bytes.Buffer
	var err error
	fmt.Printf("test")
	// If inputData is empty attempt to read from stdin
	if input == "" {
		inputData, err = root.SafeLoadStdIn()
		if err != nil {
			return instanceDetails, fmt.Errorf("failed to load stdin: %w", err)
		}
	} else {
		inputData, err = root.SafeLoadFile(input)
		if err != nil {
			return instanceDetails, fmt.Errorf("failed to load input file: %w", err)
		}
	}

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
	workflowCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&executeFlag, "exec", "e", "", "execute the WORKFLOWFILE from the direktiv package after successfully pushing.")
	workflowCmd.AddCommand(execCmd)
}
