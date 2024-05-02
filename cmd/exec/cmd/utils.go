package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const ToolName = "direktivctl"

var maxSize int64 = 1073741824

func ProjectFolder() (string, error) {
	projectFile := viper.GetString("projectFile")
	if projectFile != "" {
		return path.Dir(projectFile), nil
	}

	return "", fmt.Errorf("project directory not found")
}

func Fail(cmd *cobra.Command, s string, x ...interface{}) {
	cmd.PrintErrf(strings.TrimSuffix(s, "\n")+"\n", x...)
	os.Exit(1)
}

func pingNamespace() error {
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		UrlPrefixV2,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}

	AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping namespace %s: %w", UrlPrefixV2, err)
	}
	defer resp.Body.Close()

	// it is either ok or forbidden which means the namespace exists
	// but the user might not have access to the explorer
	// it would be still ok to e.g. send events
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		// if resp.StatusCode == http.StatusUnauthorized {

		// 	return fmt.Errorf("failed to ping namespace %s, request was unauthorized", urlGetNode)
		// }
		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to ping namespace %s, server responded with %s\n------DUMPING ERROR BODY ------\n%s", UrlPrefixV2, resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to ping namespace %s, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", UrlPrefixV2, resp.Status)
	}

	return nil
}

func SafeLoadFile(filePath string) (*bytes.Buffer, error) {
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
		return buf,
			fmt.Errorf("file is larger than maximum allowed size: %v. Set configfile 'max-size' to change",
				maxSize)
	}

	fData, err := os.ReadFile(filePath)
	if err != nil {
		return buf, err
	}

	buf = bytes.NewBuffer(fData)

	return buf, nil
}

func SafeLoadStdIn() (*bytes.Buffer, error) {
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
		return buf,
			fmt.Errorf("stdin is larger than maximum allowed size: %v. Set configfile 'max-size' to change",
				maxSize)
	}

	fData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return buf, err
	}

	buf = bytes.NewBuffer(fData)

	return buf, nil
}

func InitConfiguration(cmd *cobra.Command, args []string) {
	err := initCLI(cmd)
	if err != nil {
		Fail(cmd, "Got an error while initializing: %v", err)
	}

	cmdPrepareSharedValues()
	if err := pingNamespace(); err != nil {
		log.Fatalf("%v", err)
	}
}

func InitConfigurationAndProject(cmd *cobra.Command, args []string) {
	InitConfiguration(cmd, args)
	err := initProjectDir(cmd)
	if err != nil {
		Fail(cmd, "Got an error while initializing: %v", err)
	}
	err = InitWD()
	if err != nil {
		Fail(cmd, "Got an error while initializing: %v", err)
	}
}

func InitWD() error {
	directory := viper.GetString("directory")
	if directory == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		viper.Set("directory", pwd)
	}
	return nil
}

func GetMaxSize() int64 {
	if cfgMaxSize := viper.GetInt64("max-size"); cfgMaxSize > 0 {
		return cfgMaxSize
	}
	return maxSize
}
