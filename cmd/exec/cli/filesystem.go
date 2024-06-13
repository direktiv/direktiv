package cli

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/api"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
)

var instancesCmd = &cobra.Command{
	Use:   "filesystem",
	Short: "Execute flows and push files",
}

type instanceResponse struct {
	Data struct {
		api.InstanceData
	} `json:"data"`
}

func init() {
	RootCmd.AddCommand(instancesCmd)
	instancesCmd.AddCommand(instancesPushCmd)
	instancesCmd.AddCommand(instancesExecCmd)

	instancesExecCmd.PersistentFlags().Bool("push", true, "Push before execute.")
	// instancesExecCmd.PersistentFlags().Bool("wait", true, "Wait for flow to finish execution.")
}

var instancesExecCmd = &cobra.Command{
	Use:   "exec [name of flow]",
	Args:  cobra.ExactArgs(1),
	Short: "Execute flows in Direktiv",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := prepareCommand(cmd)
		if err != nil {
			return err
		}

		fullPath, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		projectRoot, err := findProjectRoot(fullPath)
		if err != nil {
			return err
		}

		push, err := cmd.PersistentFlags().GetBool("push")
		if err != nil {
			return err
		}

		uploader, err := newUploader(projectRoot, p)
		if err != nil {
			return err
		}

		relPath, err := GetRelativePath(projectRoot, fullPath)
		if err != nil {
			return err
		}

		// push if required
		if push {
			err = uploader.createFile(relPath, fullPath)
			if err != nil {
				return err
			}
		}

		b, err := loadStdIn()
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/api/v2/namespaces/%s/instances/?path=%s", p.Address, p.Namespace, relPath)
		resp, err := uploader.sendRequest("POST", url, b.Bytes())
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var errJson errorResponse
			err = json.Unmarshal(b, &errJson)
			if err != nil {
				return err
			}

			return fmt.Errorf(errJson.Error.Message)
		}

		id, err := handleResponse(resp, p)
		if err != nil {
			return err
		}

		err = handleOutput(p, uploader, id)
		if err != nil {
			return err
		}

		return err
	},
}

func handleOutput(profile profile, uploader *uploader, id string) error {
	fmt.Println("waiting for flow result")
	urlOutput := fmt.Sprintf("%s/api/v2/namespaces/%s/instances/%s/output", profile.Address, profile.Namespace, id)

	for i := 1; i < 20; i++ {
		resp, err := uploader.sendRequest("GET", urlOutput, nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// fetch instance id
		var instance instanceResponse
		err = json.Unmarshal(b, &instance)
		if err != nil {
			return err
		}

		if instance.Data.Status == "pending" {
			time.Sleep(1 * time.Second)
			continue
		}

		fmt.Printf("Output:\n%s\n", string(instance.Data.Output))
		break
	}

	return nil
}

func handleResponse(resp *http.Response, profile profile) (string, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// fetch instance id
	var instance instanceResponse
	err = json.Unmarshal(b, &instance)
	if err != nil {
		return "", err
	}

	fmt.Printf("executed flow with id %v\n", instance.Data.ID)

	err = printLogSSE(context.Background(), instance.Data.ID.String(), profile)
	if err != nil {
		return "", err
	}

	return instance.Data.ID.String(), nil
}

func printLogSSE(ctx context.Context, instance string, profile profile) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	urlSSE := fmt.Sprintf("%s/api/v2/namespaces/%s/logs/subscribe?instance=%s", profile.Address, profile.Namespace, instance)

	clientLogs := sse.NewClient(urlSSE)
	clientLogs.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: profile.Insecure},
	}

	if profile.Token != "" {
		clientLogs.Headers["Direktiv-Token"] = profile.Token
	}

	errCh := make(chan error, 1)

	go func() {
		err := clientLogs.SubscribeWithContext(ctx, "message", func(msg *sse.Event) {
			data := map[string]interface{}{}

			if err := json.Unmarshal(msg.Data, &data); err != nil {
				cancel()
				errCh <- err
				return
			}

			formatLogEntry(data)

			if wf, ok := data["workflow"].(map[string]interface{}); ok && (wf["status"] == string(core.LogCompletedStatus) || wf["status"] == string(core.LogFailedStatus) || wf["status"] == string(core.LogErrStatus)) {
				cancel()
				errCh <- nil
				return
			}
		})
		if err != nil {
			errCh <- err
		}
	}()

	err := <-errCh
	return err
}

func formatLogEntry(data map[string]interface{}) {
	type log struct {
		workflow interface{}
		instance interface{}
		msg      interface{}
	}

	var l log

	for key, value := range data {
		// Special handling for nested maps
		if value == nil || value == "" {
			continue
		}

		if key == "msg" {
			l.msg = value
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			wf, ok := nestedMap["workflow"]
			if ok {
				l.workflow = wf
			}
			inst, ok := nestedMap["instance"]
			if ok {
				l.instance = inst
			}
		}
	}

	fmt.Printf("%s (%v): %s\n", l.workflow, l.instance, l.msg)
}

var instancesPushCmd = &cobra.Command{
	Use:   "push [name of file/directory]",
	Args:  cobra.ExactArgs(1),
	Short: "Push files to Direktiv",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := prepareCommand(cmd)
		if err != nil {
			return err
		}

		fullPath, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		projectRoot, err := findProjectRoot(fullPath)
		if err != nil {
			return err
		}

		uploader, err := newUploader(projectRoot, p)
		if err != nil {
			return err
		}

		err = filepath.Walk(args[0], func(path string, info os.FileInfo, errIn error) error {
			fullPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			p, err := GetRelativePath(projectRoot, fullPath)
			if err != nil {
				return err
			}

			if uploader.matcher.Match(strings.Split(p, string(os.PathSeparator)), info.IsDir()) {
				cmd.Printf("skipping object %s\n", p)
				return nil
			}

			if info.IsDir() {
				err = uploader.createDirectory(p)
			} else {
				err = uploader.createFile(p, fullPath)
			}

			if err != nil {
				cmd.Printf("error creating object %s: %s\n", p, err.Error())
			}

			return nil
		})

		return err
	},
}

func GetRelativePath(configPath, targpath string) (string, error) {
	var err error

	if !filepath.IsAbs(configPath) {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return "", err
		}
	}

	if !filepath.IsAbs(targpath) {
		targpath, err = filepath.Abs(targpath)
		if err != nil {
			return "", err
		}
	}

	s, err := filepath.Rel(configPath, targpath)
	if err != nil {
		return "", err
	}

	path := filepath.ToSlash(s)
	path = strings.Trim(path, "/")

	return path, nil
}

func loadStdIn() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	fi, err := os.Stdin.Stat()
	if err != nil {
		return buf, err
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		// No stdin
		return buf, nil
	}

	fData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return buf, err
	}

	buf = bytes.NewBuffer(fData)

	return buf, nil
}
