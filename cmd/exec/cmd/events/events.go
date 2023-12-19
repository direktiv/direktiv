package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
	goutil "golang.org/x/tools/godoc/util"
)

var (
	Source      string
	Type        string
	Id          string
	Specversion string
	ContentType string
	Attachment  string
)

var eventsCmd = &cobra.Command{
	Use:              "events",
	Short:            "Event-related commands",
	PersistentPreRun: root.InitConfiguration,
}

var sendEventCmd = &cobra.Command{
	Use:   "send EVENT DATA",
	Short: "Remotely trigger direktiv events",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		urlExecuteEvent := fmt.Sprintf("%s/broadcast", root.UrlPrefix)

		filter, err := cmd.Flags().GetString("filter")
		if err != nil {
			root.Fail(cmd, "could not parse event filter %v", err.Error())
		}

		if filter != "" {
			urlExecuteEvent += "/" + strings.TrimPrefix(filter, "/")
		}

		cmd.Printf("sending events to %s\n", urlExecuteEvent)

		event, err := executeEvent(cmd, urlExecuteEvent, args)
		if err != nil {
			root.Fail(cmd, "failed to trigger event: %s %v\n", event, err)
		}

		cmd.Printf("successfully triggered event: %s\n", event)
	},
}

func executeEvent(cmd *cobra.Command, url string, args []string) (string, error) {
	event := cloudevents.NewEvent()

	// read event file in if provided
	if len(args) > 0 {
		cmd.Printf("reading cloudevent file %s\n", args[0])
		e, err := os.ReadFile(args[0])
		if err != nil {
			return "", err
		}

		// we only do json http
		err = format.Unmarshal("application/cloudevents+json", e, &event)
		if err != nil {
			return "", err
		}
	}

	// overwrite data if provided
	if Id != "" {
		event.SetID(Id)
	}

	if Specversion != "" {
		event.SetSpecVersion(Specversion)
	}

	if Source != "" {
		event.SetSource(Source)
	}

	if Type != "" {
		event.SetType(Type)
	}

	// attach data
	if len(Attachment) > 0 {
		attachment, err := os.ReadFile(Attachment)
		if err != nil {
			return "", err
		}

		// attach and guess attachment type
		ct := ContentType

		// var attach interface{}
		var attach interface{}
		err = json.Unmarshal(attachment, &attach)

		// it is not json we guess the content type if not set
		if err != nil {
			if ct == "" {
				ct = http.DetectContentType(attachment)
			}
			if goutil.IsText(attachment) {
				attach = string(attachment)
			} else {
				attach = attachment
			}
		} else {
			// if not set we assume json
			// reson for not setting it static: it could be something like whatever+json
			if ct == "" {
				ct = "application/json"
			}

			// we leave attach and use it as object
			// this converts it to json and not string json with escapes
		}

		err = event.SetData(ct, attach)
		if err != nil {
			return "", err
		}
	}

	b, err := format.JSON.Marshal(&event)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url,
		bytes.NewReader(b),
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/cloudevents+json")
	root.AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// the root command checks if the namespace exists
	// this not found has to be a wrong filter
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("eventfilter does not exist")
	} else if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("access to server forbidden")
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server responded with status %d", resp.StatusCode)
	}

	return string(b), err
}

type getFilterResp struct {
	Filtername string `json:"filtername"`
	JsCode     string `json:"jsCode"`
}

func init() {
	root.RootCmd.AddCommand(eventsCmd)
	eventsCmd.AddCommand(sendEventCmd)

	sendEventCmd.Flags().StringVar(&Attachment, "attachment", "", "Path to file used as data of the cloud event.")
	sendEventCmd.Flags().StringVar(&Source, "source", "", "Cloudevent source.")
	sendEventCmd.Flags().StringVar(&Type, "type", "", "CloudEvent type.")
	sendEventCmd.Flags().StringVar(&Id, "id", "", "Clouedevent ID. Required by spec but automatically set if not provided.")
	sendEventCmd.Flags().StringVar(&ContentType, "contentType", "", "Content type of attachment if read from file. Guessing if it is not set.")
	sendEventCmd.Flags().StringVar(&Specversion, "specversion", "", "The version of the CloudEvents specification which the event uses.")
	sendEventCmd.Flags().String("filter", "", "Custom filter for CloudEvents.")
}
