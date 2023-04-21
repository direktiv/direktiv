package events

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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

var setFilterCmd = &cobra.Command{
	Use:   "set-filter NAME SCRIPT",
	Short: "Define an event filter.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filterName := args[0]

		var (
			inputData *bytes.Buffer
			err       error
		)

		// Read input data as arg or stdin
		if len(args) > 1 {
			inputData, err = root.SafeLoadFile(args[1])
			if err != nil {
				root.Fail(cmd, "Failed to load input file: %v", err)
			}
		} else {
			inputData, err = root.SafeLoadStdIn()
			if err != nil {
				root.Fail(cmd, "Failed to load stdin: %v", err)
			}
		}

		// fail if there is nothig to create
		if inputData.Len() == 0 {
			root.Fail(cmd, "no filter function provided")
		}

		// set method to force if filter already exists
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			root.Fail(cmd, "can not read force flag: %s", err.Error())
		}

		method := http.MethodPost
		if force {
			method = http.MethodPatch
		}

		err = executeCreateCloudEventFilter(filterName, inputData, method)
		if err != nil {
			root.Fail(cmd, "can not create filter: %s\n", err.Error())
		}

		cmd.Printf("successfully created cloud event filter: %s\n", filterName)
	},
}

func executeCreateCloudEventFilter(filterName string, data io.Reader, method string) error {
	if filterName == "" {
		return errors.New("filter name not set")
	}

	url := fmt.Sprintf("%s/eventfilter/%s", root.UrlPrefix, filterName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		data,
	)
	if err != nil {
		return err
	}

	root.AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// this can not happen on update
	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("event filter %s already exists", filterName)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("event filter %s invalid. check syntax", filterName)
	}

	return nil
}

type listFiltersResp struct {
	EventFilter []struct {
		Name string `json:"name"`
	} `json:"eventFilter"`
}

var listFilterCmd = &cobra.Command{
	Use:   "list-filters",
	Short: "List event filters for namespace.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := executeListCloudEventFilter()
		if err != nil {
			root.Fail(cmd, "can not fetch event filter: %v\n", err)
		}

		var eventfilter listFiltersResp
		err = json.Unmarshal(resp, &eventfilter)
		if err != nil {
			root.Fail(cmd, "can not unmarshall event filter response: %v\n", err)
		}

		for i := range eventfilter.EventFilter {
			cmd.Println(eventfilter.EventFilter[i].Name)
		}
	},
}

func executeListCloudEventFilter() ([]byte, error) {
	var err error

	url := fmt.Sprintf("%s/eventfilter", root.UrlPrefix)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
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
		err = fmt.Errorf("failed to list filters (rejected by server)")
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}

var deleteFilterCmd = &cobra.Command{
	Use:   "delete-filter NAME",
	Short: "Delete an event filter.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filterName := args[0]

		err := executeDeleteCloudEventFilter(filterName)
		if err != nil {
			root.Fail(cmd, "error: %v\n", err)
		}

		cmd.Printf("successfully deleted cloud event filter: %s\n", filterName)
	},
}

func executeDeleteCloudEventFilter(filterName string) error {
	var err error

	if filterName == "" {
		err = fmt.Errorf("filtername was not set")
		return err
	}

	url := fmt.Sprintf("%s/eventfilter/%s", root.UrlPrefix, filterName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		url,
		nil,
	)
	if err != nil {
		return err
	}

	root.AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		err = fmt.Errorf("filter " + filterName + " does not exist")
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to delete filter: %s (error code: %d)", filterName, resp.StatusCode)
		return err
	}

	return err
}

type getFilterResp struct {
	Filtername string `json:"filtername"`
	JsCode     string `json:"jsCode"`
}

var getFilterCmd = &cobra.Command{
	Use:   "get-filter NAME",
	Short: "Get an event filter.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filterName := args[0]

		resp, err := executeGetCloudEventFilter(filterName)
		if err != nil {
			root.Fail(cmd, "error: %v\n", err)
		}

		var eventfilter getFilterResp
		err = json.Unmarshal(resp, &eventfilter)
		if err != nil {
			root.Fail(cmd, "error: %v\n", err)
		}

		cmd.Printf("filtername: %s\n", eventfilter.Filtername)
		cmd.Printf("script: %s\n", eventfilter.JsCode)
	},
}

func executeGetCloudEventFilter(filterName string) ([]byte, error) {
	var err error

	if filterName == "" {
		return nil, fmt.Errorf("filter name was not set")
	}

	url := fmt.Sprintf("%s/eventfilter/%s", root.UrlPrefix, filterName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	root.AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		err = fmt.Errorf("filter " + filterName + " does not exist")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get filter: %s (error code: %d)", filterName, resp.StatusCode)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	return body, err
}

func init() {
	root.RootCmd.AddCommand(eventsCmd)
	eventsCmd.AddCommand(sendEventCmd)
	eventsCmd.AddCommand(setFilterCmd)
	eventsCmd.AddCommand(deleteFilterCmd)
	eventsCmd.AddCommand(getFilterCmd)
	eventsCmd.AddCommand(listFilterCmd)

	sendEventCmd.Flags().StringVar(&Attachment, "attachment", "", "Path to file used as data of the cloud event.")
	sendEventCmd.Flags().StringVar(&Source, "source", "", "Cloudevent source.")
	sendEventCmd.Flags().StringVar(&Type, "type", "", "CloudEvent type.")
	sendEventCmd.Flags().StringVar(&Id, "id", "", "Clouedevent ID. Required by spec but automatically set if not provided.")
	sendEventCmd.Flags().StringVar(&ContentType, "contentType", "", "Content type of attachment if read from file. Guessing if it is not set.")
	sendEventCmd.Flags().StringVar(&Specversion, "specversion", "", "The version of the CloudEvents specification which the event uses.")
	sendEventCmd.Flags().String("filter", "", "Custom filter for CloudEvents.")

	setFilterCmd.PersistentFlags().BoolP("force", "f", false, "Forced update for event filter if it already exists.")
}
