package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/api"
)

// DirektivURL stroes the endpoint for direktiv
var DirektivURL string

// Content types for requests
const (
	JSONCt = "application/json"
	YAMLCt = "text/yaml"
	NONECt = ""
)

// GenerateCmd is a basic cobra function
func GenerateCmd(use, short, long string, fn func(cmd *cobra.Command, args []string), c cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run:   fn,
		Args:  c,
	}
}

// DoRequest executes request against ingress
func DoRequest(method, path, ct string, body *string) ([]byte, error) {

	var out []byte

	d := time.Now().Add(10 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	var rd io.Reader
	if body != nil {
		rd = strings.NewReader(*body)
	}

	req, err := http.NewRequestWithContext(ctx, method,
		fmt.Sprintf("%s/api%s", DirektivURL, path), rd)
	if err != nil {
		return out, err
	}

	if len(ct) > 0 {
		req.Header.Set("Content-type", ct)
	}

	c := &http.Client{}

	res, err := c.Do(req)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()

	out, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	if res.StatusCode != 200 {
		var eo api.ErrObject
		err := json.Unmarshal(out, &eo)
		if err != nil {
			log.Fatalf("can not parse error response: %v", err)
		}
		return out, fmt.Errorf(eo.Message)
	}

	return out, nil

}

// CLILogWriter for using fatalf without date
type CLILogWriter struct {
}

func (writer CLILogWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}
