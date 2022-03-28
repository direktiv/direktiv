package main

import (
	"crypto/tls"
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

const DefaultConfigName = ".direktiv.conf"

// Flags
var (
	addr       string
	path       string
	input      string
	inputType  string
	outputFlag string
	namespace  string

	apiKey    string
	authToken string
	insecure  bool

	maxSize int64 = 1073741824

	configPath string
)

// Shared Vars
var (
	configPathFromFlag bool = true
	localAbsPath       string
	urlPrefix          string
	urlWorkflow        string
	urlUpdateWorkflow  string
)

func main() {

	var err error

	// Read Config
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Loads flag values from YAML config if file is found. If unset will automtically look for config file in workflow yaml and parent directories")

	// Load config flag early
	loadCfgFlag()

	// Walk Up to search for config
	if configPath == "" {
		autoConfigPathFinder()
	}

	viper.SetConfigType("yml")
	viper.SetConfigFile(configPath)
	viper.ReadInConfig()

	// Set Flags
	rootCmd.Flags().StringP("addr", "a", "", "Target direktiv api address. "+configFlagHelpTextLoader("addr", false))

	execCmd.Flags().StringP("path", "p", "", "Target remote workflow path .e.g. '/dir/workflow'. Automatically set if config file was auto-set. "+configFlagHelpTextLoader("path", false))
	pushCmd.Flags().StringP("path", "p", "", "Target remote path. e.g. '/dir/workflow'. Automatically set if config file was auto-set. If pushing local dir config flag cannot be used. "+configFlagHelpTextLoader("path", false))

	execCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Path where to write instance output. If unset output will be written to screen")
	execCmd.Flags().StringVarP(&input, "input", "i", "", "Path to file to be used as input data for executed workflow. If unset, stdin will be used as input data if available.")
	execCmd.Flags().StringVar(&inputType, "input-type", "application/json", "Content Type of input data")
	rootCmd.Flags().StringP("namespace", "n", "", "Target namespace to execute workflow on. "+configFlagHelpTextLoader("namespace", false))
	rootCmd.Flags().StringP("api-key", "k", "", "Authenticate request with apikey. "+configFlagHelpTextLoader("api-key", true))
	rootCmd.Flags().StringP("auth-token", "t", "", "Authenticate request with token. "+configFlagHelpTextLoader("auth-token", true))
	rootCmd.Flags().BoolVar(&insecure, "insecure", true, "Accept insecure https connections")

	// Bing CLI flags to viper
	configBindFlag(rootCmd, "addr", true)

	// If config was automatically found, path is no longer required
	configBindFlag(execCmd, "path", configPathFromFlag)
	configBindFlag(pushCmd, "path", configPathFromFlag)
	configBindFlag(rootCmd, "namespace", true)
	configBindFlag(rootCmd, "api-key", false)
	configBindFlag(rootCmd, "auth-token", false)

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
	addr = viper.GetString("addr")
	path = viper.GetString("path")
	namespace = viper.GetString("namespace")
	apiKey = viper.GetString("api-key")
	authToken = viper.GetString("auth-token")
	if cfgMaxSize := viper.GetInt64("max-size"); cfgMaxSize > 0 {
		maxSize = cfgMaxSize
	}

	// Get ABS Path
	localAbsPath, err = filepath.Abs(wfPath)
	if err != nil {
		log.Fatalf("Failed to locate workflow file in filesystem: %v\n", err)
	}

	// If config file was found automatically, generate path relative to config dir
	if !configPathFromFlag {
		os.Stderr.WriteString(fmt.Sprintf("Using config file: '%s'\n", configPath))
		path = strings.TrimSuffix(strings.TrimPrefix(localAbsPath, filepath.Dir(configPath)), ".yaml")
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Using flag config file: '%s'\n", configPath))
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure}

	urlPrefix = fmt.Sprintf("%s/api/namespaces/%s", addr, namespace)
	urlWorkflow = fmt.Sprintf("%s/tree/%s", urlPrefix, strings.TrimPrefix(path, "/"))
	urlUpdateWorkflow = fmt.Sprintf("%s?op=update-workflow", urlWorkflow)
}

var rootCmd = &cobra.Command{
	Use: "exec",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmdPrepareWorkflow(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		pathsToUpdate := make([]string, 0)
		pathStat, err := os.Stat(localAbsPath)
		if err != nil {
			log.Fatalf("Could not access path: %v", err)
		}
		if pathStat.IsDir() {
			if configPathFromFlag {
				log.Fatal("Config file must be automatically found when push directory")
			}

			err = filepath.Walk(localAbsPath,
				func(localPath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if strings.HasSuffix(localPath, ".yaml") {
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
			path = strings.TrimSuffix(strings.TrimPrefix(localPath, filepath.Dir(configPath)), ".yaml")
			urlWorkflow = fmt.Sprintf("%s/tree/%s", urlPrefix, strings.TrimPrefix(path, "/"))
			urlUpdateWorkflow = fmt.Sprintf("%s?op=update-workflow", urlWorkflow)

			cmd.PrintErrf("[%v/%v] Updating Namespace: '%s' Workflow: '%s'\n", i+1, len(pathsToUpdate), namespace, path)
			err = updateRemoteWorkflow(urlUpdateWorkflow, localPath)
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
				varName := strings.TrimPrefix(v, localPath+".")
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmdPrepareWorkflow(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		instanceStatus := "pending"

		cmd.PrintErrf("Updating Namespace: '%s' Workflow: '%s'\n", namespace, path)
		err := updateRemoteWorkflow(urlUpdateWorkflow, localAbsPath)
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
			varName := strings.TrimPrefix(v, localAbsPath+".")
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
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		}

		clientLogs.Headers["apikey"] = apiKey
		clientLogs.Headers["Direktiv-Token"] = authToken

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
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		}

		clientInstance.Headers["apikey"] = apiKey
		clientInstance.Headers["Direktiv-Token"] = authToken

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
