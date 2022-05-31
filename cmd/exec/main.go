package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flags
var (
	input      string
	inputType  string
	outputFlag string

	maxSize int64 = 1073741824
)

// Shared Vars
var (
	localAbsPath string
	urlPrefix    string
	urlWorkflow  string
)

func main() {

	// Read Config
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(pushCmd)

	rootCmd.PersistentFlags().StringP("profile", "P", "", "Select the named profile from the loaded multi-profile configuration file.")
	rootCmd.PersistentFlags().StringP("directory", "C", "", "Change to this directory before evaluating any paths or searching for a configuration file.")

	rootCmd.PersistentFlags().StringP("addr", "a", "", "Target direktiv api address.")
	rootCmd.PersistentFlags().StringP("path", "p", "", "Target remote workflow path .e.g. '/dir/workflow'. Automatically set if config file was auto-set.")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Target namespace to execute workflow on.")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "Authenticate request with apikey.")
	rootCmd.PersistentFlags().StringP("auth-token", "t", "", "Authenticate request with token.")
	rootCmd.PersistentFlags().Bool("insecure", true, "Accept insecure https connections")

	execCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Path where to write instance output. If unset output will be written to screen")
	execCmd.Flags().StringVarP(&input, "input", "i", "", "Path to file to be used as input data for executed workflow. If unset, stdin will be used as input data if available.")
	execCmd.Flags().StringVar(&inputType, "input-type", "application/json", "Content Type of input data")

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		fail("error binding configuration flags: %v", err)
	}

	viper.SetEnvPrefix("direktiv")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	err = rootCmd.Execute()
	if err != nil {
		log.Fatalf("Command Failed: %v", err)
	}

}

func getOutput(url string) ([]byte, error) {
	var output instanceOutput

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	addAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
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

func cmdPrepareWorkflow(wfPath string) {
	var err error

	// Load Config From flags / config
	addr := getAddr()
	namespace := getNamespace()

	if cfgMaxSize := viper.GetInt64("max-size"); cfgMaxSize > 0 {
		maxSize = cfgMaxSize
	}

	// Get ABS Path
	localAbsPath, err = filepath.Abs(wfPath)
	if err != nil {
		log.Fatalf("Failed to locate workflow file in filesystem: %v\n", err)
	}

	path := getPath(wfPath)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = getTLSConfig()

	urlPrefix = fmt.Sprintf("%s/api/namespaces/%s", addr, namespace)
	urlWorkflow = fmt.Sprintf("%s/tree/%s", urlPrefix, strings.TrimPrefix(path, "/"))
}

var rootCmd = &cobra.Command{
	Use: "exec",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		loadConfig(cmd)

	},
}

var pushCmd = &cobra.Command{
	Use:   "push WORKFLOW_PATH|DIR_PATH",
	Short: "Pushes local workflow or dir to remote direktiv server. This process will update your latest remote resource to your local WORKFLOW_PATH|DIR_PATH file",
	Long: `"Pushes local workflow or dir to remote direktiv server. This process will update your latest remote resource to your local WORKFLOW_PATH|DIR_PATH file.
Pushing local directory cannot be used with config flag. Config must be found automatically to determine folder structure.

EXAMPLE: push helloworld.yaml --addr http://192.168.1.1 --namespace admin

Variables will also be uploaded if they are prefixed with your local workflow name
EXMAPLE:  
  dir: /pwd
        /helloworld.yaml
        /helloworld.yaml.data.json
Executing: push helloworld.yaml --addr http://192.168.1.1 --namespace admin --path helloworld
Will update the helloworld workflow and set the remote workflow variable 'data.json' to the contents of '/helloworld.yaml.data.json'
`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmdPrepareWorkflow(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		pathsToUpdate := make([]string, 0)
		pathStat, err := os.Stat(localAbsPath)
		if err != nil {
			log.Fatalf("Could not access path: %v", err)
		}
		if pathStat.IsDir() {
			err = filepath.Walk(localAbsPath,
				func(localPath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if (strings.HasSuffix(localPath, ".yaml") || strings.HasSuffix(localPath, ".yml")) && !(strings.Contains(localPath, ".yaml.") || strings.Contains(localPath, ".yml.")) {
						pathsToUpdate = append(pathsToUpdate, localPath)
					}
					return nil
				})

			if err != nil {
				log.Fatalf("Recursive search could not access path: %v", err)
			}
		} else {
			pathsToUpdate = append(pathsToUpdate, localAbsPath)
		}

		cmd.PrintErrf("Found %v Local Workflow/s to update\n", len(pathsToUpdate))
		for i, localPath := range pathsToUpdate {
			path := getPath(localPath)

			cmd.PrintErrf("[%v/%v] Updating Namespace: '%s' Workflow: '%s'\n", i+1, len(pathsToUpdate), getNamespace(), path)
			err = updateRemoteWorkflow(path, localPath)
			if err != nil {
				log.Fatalf("Failed to update remote workflow: %v\n", err)
			}

			localVars, err := getLocalWorkflowVariables(localPath)
			if err != nil {
				log.Fatalf("Failed to get local variable files: %v\n", err)
			}
			if len(localVars) > 0 {
				cmd.PrintErrf("Found %v Local Variables to push to remote\n", len(localVars))
			}

			// Set Remote Vars
			for _, v := range localVars {
				varName := filepath.ToSlash(strings.TrimPrefix(v, localPath+"."))
				cmd.PrintErrf("      Updating Remote Workflow Variable: '%s'\n", varName)
				err = setRemoteWorkflowVariable(urlWorkflow, varName, v)
				if err != nil {
					log.Fatalf("Failed to set remote variable file: %v\n", err)
				}
			}

			cmd.PrintErrf("      Successfully updated remote workflow\n")
		}
	},
}

var execCmd = &cobra.Command{
	Use:   "exec WORKFLOW_PATH",
	Short: "Remotely execute direktiv workflows with local files. This process will update your latest remote workflow to your local WORKFLOW_PATH file",
	Long: `Remotely execute direktiv workflows with local files. This process will update your latest remote workflow to your local WORKFLOW_PATH file.

EXAMPLE: exec helloworld.yaml --addr http://192.168.1.1 --namespace admin --path helloworld

Variables will also be uploaded if they are prefixed with your local workflow name
EXMAPLE:  
  dir: /pwd
        /helloworld.yaml
        /helloworld.yaml.data.json
Executing: exec helloworld.yaml --addr http://192.168.1.1 --namespace admin --path helloworld
Will update the helloworld workflow and set the remote workflow variable 'data.json' to the contents of '/helloworld.yaml.data.json'
`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		cmdPrepareWorkflow(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		instanceStatus := "pending"

		path := getPath(args[0])

		err := updateRemoteWorkflow(path, localAbsPath)
		if err != nil {
			log.Fatalf("Failed to update remote workflow: %v\n", err)
		}

		localVars, err := getLocalWorkflowVariables(localAbsPath)
		if err != nil {
			log.Fatalf("Failed to get local variable files: %v\n", err)
		}
		if len(localVars) > 0 {
			cmd.PrintErrf("Found %v Local Variables to push to remote\n", len(localVars))
		}

		// Set Remote Vars
		for _, v := range localVars {
			varName := filepath.ToSlash(strings.TrimPrefix(v, localAbsPath+"."))
			cmd.PrintErrf("Updating Remote Workflow Variable: '%s'\n", varName)
			err = setRemoteWorkflowVariable(urlWorkflow, varName, v)
			if err != nil {
				log.Fatalf("Failed to set remote variable file: %v\n", err)
			}
		}

		urlExecute := fmt.Sprintf("%s/tree/%s?op=execute&ref=latest", urlPrefix, strings.TrimPrefix(path, "/"))
		instanceDetails, err := executeWorkflow(urlExecute)
		if err != nil {
			log.Fatalf("Failed to execute workflow: %v\n", err)
		}

		cmd.PrintErrf("Successfully Executed Instance: %s\n", instanceDetails.Instance)
		cmd.PrintErrln("-------INSTANCE LOGS-------")
		urlLogs := fmt.Sprintf("%s/instances/%s/logs", urlPrefix, instanceDetails.Instance)
		clientLogs := sse.NewClient(urlLogs)
		clientLogs.Connection.Transport = &http.Transport{
			TLSClientConfig: getTLSConfig(),
		}

		addSSEAuthHeaders(clientLogs)

		logsChannel := make(chan *sse.Event)
		clientLogs.SubscribeChan("messages", logsChannel)

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

				if len(logResp.Edges) > 0 {
					for _, edge := range logResp.Edges {
						cmd.PrintErrf("%v: %s\n", edge.Node.T.In(time.Local).Format("02 Jan 06 15:04 MST"), edge.Node.Msg)
					}
				}
			}
		}()

		urlInstance := fmt.Sprintf("%s/instances/%s", urlPrefix, instanceDetails.Instance)
		clientInstance := sse.NewClient(urlInstance)
		clientInstance.Connection.Transport = &http.Transport{
			TLSClientConfig: getTLSConfig(),
		}

		addSSEAuthHeaders(clientInstance)

		channelInstance := make(chan *sse.Event)
		clientInstance.SubscribeChan("messages", channelInstance)
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

		cmd.PrintErrf("Instance Completed With Status: %s\n", instanceStatus)
		urlOutput := fmt.Sprintf("%s/instances/%s/output", urlPrefix, instanceDetails.Instance)

		output, err := getOutput(urlOutput)
		if outputFlag != "" {
			err = os.WriteFile(outputFlag, output, 0644)
			if err != nil {
				log.Fatalf("Failed to write output file: %v\n", err)
			}
		} else {
			cmd.PrintErrln("------INSTANCE OUTPUT------")
			fmt.Println(string(output))
		}
	},
}

func fail(s string, x ...interface{}) {

	fmt.Fprintf(os.Stderr, strings.TrimSuffix(s, "\n")+"\n", x...)
	os.Exit(1)

}

func printlog(s string, x ...interface{}) {
	fmt.Fprintf(os.Stderr, strings.TrimSuffix(s, "\n")+"\n", x...)
}
