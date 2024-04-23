package workflows

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
	Short: "Workflow-related commands. Execute commands of this palette from a direktiv project directory, or a related sub directory.",
	Long: `Use this command palette from a direktiv project directory, or a related sub directory. 
A direktiv-project can contain following files:

- .direktiv.yaml:
REQUIRED: A File at the top directory of the project. This file can contain regex-rules to ignore unrelated files in the project directory from the project.

- subdirectories.

- direktiv-workflows:
A direktiv workflow have to be a .yaml file. Keep in mind that the workflow-name on the server will preserve the file-extension as part of the workflow-name

- workflows-variables:
A workflows variable has to carry the workflow name as a prefix and the variable name as a sufix. For example:
	- workflows name: start.yaml
	- variable name: start.yaml.data.json 
`,
	PersistentPreRun: root.InitConfigurationAndProject,
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

var pushCmd = &cobra.Command{
	Use:   "push WORKFLOW_PATH|DIR_PATH",
	Short: "Adds new or updates workflows and variables to the namespace of the remove server",
	Long: `Adds new or updates workflows and variables to the namespace of the remove server. Run this command from a direktiv-project-folder. Example: 
	
	push helloworld.yaml --addr http://192.168.1.1 --namespace admin

When a folder is passed as an argument all the resources in the specified folder will be created or updated in the same directory-structure as the local direktiv-project.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pathsToUpdate, err := getImpactedFiles(args[0], true, true)
		if err != nil {
			root.Fail(cmd, "could not calculate impacted files: %v", err)
		}

		relativeDir := root.GetProjectFileLocation()

		// cull using ignores
		x := make([]string, 0)
		for i := range pathsToUpdate {

			p := pathsToUpdate[i]
			path, err := root.GetRelativePath(relativeDir, p)
			if err != nil {
				root.Fail(cmd, "Could not calculate path, %v", err)
			}

			shouldAdd := true
			for _, g := range root.Globbers {
				if g.Match(path) {
					cmd.Printf("ignoring workflow %s\n", path)
					shouldAdd = false
					break
				}
			}

			if shouldAdd {
				x = append(x, p)
			}
		}

		cmd.Printf("found %v local workflow/s to update\n", len(x))

		for i := range x {
			wf := x[i]
			path, err := root.GetRelativePath(relativeDir, wf)
			if err != nil {
				root.Fail(cmd, "Could not calculate paths %v", err)
			}

			cmd.Printf("pushing workflow %s\n", path)
			cmd.Printf("updating namespace: '%s' workflow: '%s'\n", root.GetNamespace(), path)

			err = updateRemoteWorkflow(path, wf)
			if err != nil {
				fmt.Printf("can not update workflow: %s\n", err.Error())
			}
		}
	},
}

func updateRemoteWorkflow(path string, localPath string) error {
	err := recurseMkdirParent(path)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
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
		defer resp.Body.Close()

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
		urlDir := fmt.Sprintf("%s/files/%s", root.UrlPrefixV2, strings.Trim(createPath, "/"))

		jsonData, err := json.Marshal(struct {
			Name string `json:"name"`
			Typ  string `json:"type"`
		}{
			Name: strings.Trim(createPath, "/"),
			Typ:  "directory",
		})
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)

			return fmt.Errorf("failed to create request file: %w", err)
		}

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			urlDir,
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return fmt.Errorf("failed to create request file: %w", err)
		}

		root.AddAuthHeaders(req)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

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

var (
	outputFlag     string
	execNoPushFlag bool
	input          string
)

var execCmd = &cobra.Command{
	Use:   "exec WORKFLOW_PATH",
	Short: "Uploads and executes a direktiv workflow for a direktiv project.",
	Long: `Uploads and executes a direktiv workflow for a direktiv project. EXAMPLE: 
	
	exec helloworld.yaml --addr http://192.168.1.1 --namespace admin

The Workflow can be executed with input data by passing it via stdin or the input flag.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		relativeDir := root.GetProjectFileLocation()
		path, err := root.GetRelativePath(relativeDir, args[0])
		if err != nil {
			root.Fail(cmd, "Failed to calculate path %v", err)
		}

		if !execNoPushFlag {
			err := updateRemoteWorkflow(path, args[0])
			if err != nil {
				cmd.Printf("can not execute workflow: %v\n", err)
			}
		} else {
			cmd.Printf("skipping updating namespace: '%s' workflow: '%s'\n", root.GetNamespace(), path)
		}

		urlExecute := fmt.Sprintf("%s/tree/%s?op=execute&ref=latest", root.UrlPrefix, strings.TrimPrefix(path, "/"))
		instanceDetails, err := executeWorkflow(urlExecute)
		if err != nil {
			root.Fail(cmd, "Failed to execute workflow: %v\n", err)
		}
		cmd.Printf("Successfully Executed Instance: %s\n", instanceDetails.Instance)
		urlOutput := fmt.Sprintf("%s/instances/%s/output", root.UrlPrefix, instanceDetails.Instance)
		urlsse := fmt.Sprintf("%s/logs/subscribe?instance=%s", root.UrlPrefixV2, instanceDetails.Instance) // Construct SSE log subscription URL
		out := func(msg string) {
			cmd.Println(msg)
		}
		err = root.GetLogsSSE(cmd.Context(), out, urlsse) // Use the refactored function
		if err != nil {
			cmd.PrintErr("Error fetching logs: ", err)
		}

		output, err := getOutput(urlOutput)
		if err != nil {
			root.Fail(cmd, "%s", err)
		}
		if outputFlag != "" {
			err := os.WriteFile(outputFlag, output, 0o600)
			if err != nil {
				log.Fatalf("failed to write output file: %v\n", err)
			}
		} else {
			cmd.Println("------INSTANCE OUTPUT------")
			cmd.Println(string(output))
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
		return instanceDetails, fmt.Errorf("failed to load input file: %w", err)
	}

	// If inputData is empty attempt to read from stdin
	if inputData.Len() == 0 {
		inputData, err = root.SafeLoadStdIn()
		if err != nil {
			return instanceDetails, fmt.Errorf("failed to load stdin: %w", err)
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	workflowCmd.PersistentFlags().StringP("directory", "C", "", "Runs the command as if "+root.ToolName+" was started in the given directory instead of the current working directory.")

	workflowCmd.AddCommand(pushCmd)
	workflowCmd.AddCommand(execCmd)

	execCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Path where to write instance output. If unset output will be written to screen")
	execCmd.Flags().StringVarP(&input, "input", "i", "", "Path to file to be used as input data for executed workflow. If unset, stdin will be used as input data if available.")
	execCmd.Flags().BoolVar(&execNoPushFlag, "no-push", false, "If set will skip updating and just execute the workflow.")

	workflowCmd.AddCommand(infoCmd)
	infoCmd.PersistentFlags().StringP("directory", "C", "", "Runs the command as if "+root.ToolName+" was started in the given directory instead of the current working directory.")
}
