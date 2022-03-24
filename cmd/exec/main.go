package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
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

// Manually load config flag
func loadCfgFlag() {
	flag.Parse()
	var foundFlag bool
	for _, arg := range flag.Args() {
		if foundFlag {
			configPath = arg
			break
		}

		if arg == "--config" || arg == "-c" {
			foundFlag = true
			continue
		}

		if strings.HasPrefix(arg, "-c=") {
			configPath = strings.TrimPrefix(arg, "-c=")
			break
		}

		if strings.HasPrefix(arg, "--config=") {
			configPath = strings.TrimPrefix(arg, "--config=")
			break
		}
	}
}

func initDefaultConfigPath() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Warning: Failed to get home directory. Could not establish config file for defaults: %s\n", err.Error()))
		return ""
	}
	configHome := dirname
	configName := ".dre-config"
	configType := "yml"
	return filepath.Join(configHome, configName+"."+configType)
}

// configFlagHelpTextLoader : Generate suffix for flag help text to show set config value.
func configFlagHelpTextLoader(configKey string, sensitive bool) (flagHelpText string) {
	configValue := viper.GetString(configKey)

	if configValue != "" {
		if sensitive {
			flagHelpText = "(config \"***************\")"
		} else {
			flagHelpText = fmt.Sprintf("(config \"%s\")", configValue)
		}
	}

	return
}

//	configBindFlag : Binds cli flag for config value. If flag value is set, will be used instead of config value.
//	If config value is not set, mark flag as required.
func configBindFlag(configKey string, required bool) {
	viper.BindPFlag(configKey, rootCmd.Flags().Lookup(configKey))
	if required && viper.GetString(configKey) == "" {
		rootCmd.MarkFlagRequired(configKey)
	}
}

func addAuthHeaders(req *http.Request) {
	req.Header.Add("apikey", apiKey)
	req.Header.Add("Direktiv-Token", authToken)
}

func safeLoadFile(filePath string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	if filePath == "" {
		// skip if filePath is empty
		return buf, nil
	}

	fStat, err := os.Stat(filePath)
	if err != nil {
		return buf, err
	}

	if fStat.Size() > maxSize {
		return buf, fmt.Errorf("file is larger than maximum allowed size: %v. Set configfile 'max-size' to change", maxSize)
	}

	fData, err := os.ReadFile(filePath)
	if err != nil {
		return buf, err
	}

	buf = bytes.NewBuffer(fData)

	return buf, nil
}

func safeLoadStdIn() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	fi, err := os.Stdin.Stat()
	if err != nil {
		return buf, err
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		// No stdin
		return buf, nil
	}

	if fi.Size() > maxSize {
		return buf, fmt.Errorf("stdin is larger than maximum allowed size: %v. Set configfile 'max-size' to change", maxSize)
	}

	fData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return buf, err
	}

	buf = bytes.NewBuffer(fData)

	return buf, nil
}

func main() {

	var err error

	// Read Config
	rootCmd.Flags().StringVarP(&configPath, "config", "c", initDefaultConfigPath(), "Loads flag values from YAML config if file is found.")

	// Load config flag early
	loadCfgFlag()

	viper.SetConfigFile(configPath)
	viper.ReadInConfig()

	// Set Flags
	rootCmd.Flags().StringP("addr", "a", "", "Target direktiv api address. "+configFlagHelpTextLoader("addr", false))
	rootCmd.Flags().StringP("path", "p", "", "Target remote workflow path .e.g. '/dir/workflow'. "+configFlagHelpTextLoader("path", false))
	rootCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Path where to write instance output. If unset output will be written to screen")
	rootCmd.Flags().StringVarP(&input, "input", "i", "", "Path to file to be used as input data for executed workflow. If unset, stdin will be used as input data if available.")
	rootCmd.Flags().StringVar(&inputType, "input-type", "application/json", "Content Type of input data")
	rootCmd.Flags().StringP("namespace", "n", "", "Target namespace to execute workflow on. "+configFlagHelpTextLoader("namespace", false))
	rootCmd.Flags().StringP("api-key", "k", "", "Authenticate request with apikey. "+configFlagHelpTextLoader("api-key", true))
	rootCmd.Flags().StringP("auth-token", "t", "", "Authenticate request with token. "+configFlagHelpTextLoader("auth-token", true))
	rootCmd.Flags().BoolVar(&insecure, "insecure", true, "Accept insecure https connections")

	// Bing CLI flags to viper
	configBindFlag("addr", true)
	configBindFlag("path", true)
	configBindFlag("namespace", true)
	configBindFlag("api-key", false)
	configBindFlag("auth-token", false)

	err = rootCmd.Execute()
	if err != nil {
		log.Fatalf("Command Failed: %v", err)
	}

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

func updateRemoteWorkflow(url string, localPath string) error {
	wfData, err := safeLoadFile(localPath)
	if err != nil {
		log.Fatalf("Failed to load workflow file: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		wfData,
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
			return fmt.Errorf("failed to update workflow, server responsed with %s\n------DUMPING ERROR BODY ------\n%s", resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to update workflow, server responsed with %s\n------DUMPING ERROR BODY ------\nCould read response body", resp.Status)
	}

	return err
}

var rootCmd = &cobra.Command{
	Use:   "exec WORKFLOW_PATH",
	Short: "Remotely execute direktiv workflows with local files. This process will update your latest remote workflow to your local WORKFLOW_PATH file",
	Long: `Remotely execute direktiv workflows with local files. This process will update your latest remote workflow to your local WORKFLOW_PATH file.

EXAMPLE: exec helloworld.yaml --addr http://192.168.1.1 --namespace admin --path helloworld`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load Config From flags / config
		addr = viper.GetString("addr")
		path = viper.GetString("path")
		namespace = viper.GetString("namespace")
		apiKey = viper.GetString("api-key")
		authToken = viper.GetString("auth-token")
		if cfgMaxSize := viper.GetInt64("max-size"); cfgMaxSize > 0 {
			maxSize = cfgMaxSize
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure}

		instanceStatus := "pending"
		urlPrefix := fmt.Sprintf("%s/api/namespaces/%s", addr, namespace)
		urlUpdateWorkflow := fmt.Sprintf("%s/tree/%s?op=update-workflow", urlPrefix, strings.TrimPrefix(path, "/"))

		cmd.PrintErrf("Updating Namespace: '%s' Workflow: '%s'\n", namespace, path)
		err := updateRemoteWorkflow(urlUpdateWorkflow, args[0])
		if err != nil {
			log.Fatalf("Failed to update remote workflow: %v\n", err)
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
