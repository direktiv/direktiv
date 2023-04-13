package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var maxSize int64 = 1073741824

func ConfigFilePath() (string, error) {
	cd := "unset"

	if config.path != "" {
		return config.path, nil
	}

	return cd, nil
}

func ProjectFolder() (string, error) {
	pd := config.ProfileConfig.Path
	if pd == "" {
		return "path is not set", fmt.Errorf("path is not set in the config-file")
	}

	return pd, nil
}

func Fail(s string, x ...interface{}) {
	fmt.Fprintf(os.Stderr, strings.TrimSuffix(s, "\n")+"\n", x...)
	os.Exit(1)
}

func Printlog(s string, x ...interface{}) {
	fmt.Fprintf(os.Stderr, strings.TrimSuffix(s, "\n")+"\n", x...)
}

func pingNamespace() error {
	urlGetNode := fmt.Sprintf("%s/tree/", UrlPrefix)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		urlGetNode,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}

	AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping namespace %s: %w", urlGetNode, err)
	}

	// it is either ok or forbidden which means the namespace exists
	// but the user might not have access to the explorer
	// it would be still ok to e.g. send events
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		// if resp.StatusCode == http.StatusUnauthorized {

		// 	return fmt.Errorf("failed to ping namespace %s, request was unauthorized", urlGetNode)
		// }
		errBody, err := io.ReadAll(resp.Body)
		if err == nil {
			return fmt.Errorf("failed to ping namespace %s, server responded with %s\n------DUMPING ERROR BODY ------\n%s", urlGetNode, resp.Status, string(errBody))
		}

		return fmt.Errorf("failed to ping namespace %s, server responded with %s\n------DUMPING ERROR BODY ------\nCould read response body", urlGetNode, resp.Status)
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
