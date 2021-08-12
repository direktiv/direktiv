package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
	tailfile "github.com/nxadm/tail"
	log "github.com/sirupsen/logrus"
	_ "github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/flow"
)

type errStruct struct {
	Code    string
	Message string
}

func readErrorFile() (code, msg string, err error) {

	a, err := ioutil.ReadFile("/direktiv-data/error.json")
	if err != nil {
		return "", "", err
	}

	es := new(errStruct)

	err = json.Unmarshal(a, es)
	if err != nil {
		return "", "", err
	}

	return es.Code, es.Message, nil

}

func runAsSidecar() {

	log.Println("Running as sidecar container.")

	go tail()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {

		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				log.Printf("Inotify event: %v", event)

				if event.Op&fsnotify.Create == fsnotify.Create && event.Name == "/direktiv-data/done" {

					log.Println("\"Done\" file created.")

					err := setOutVariables(context.Background())
					if err != nil {
						log.Printf("Error setting output file variables: %v", err)
					}

					output, err := ioutil.ReadFile("/direktiv-data/output.json")
					if err != nil {
						log.Printf("Error reading output file: %v", err)
					}

					errCode, errMsg, err := readErrorFile()
					if err != nil {
						log.Printf("Error reading error file: %v", err)
					}

					_, err = flowClient.ReportActionResults(context.Background(), &flow.ReportActionResultsRequest{
						InstanceId:   &instanceId,
						Step:         &step,
						ActionId:     &actionId,
						ErrorCode:    &errCode,
						ErrorMessage: &errMsg,
						Output:       output,
					})
					if err != nil {
						log.Printf("Error reporting action results: %v", err)
						return
					}

					// TODO: file vars
					log.Printf("Job finished.")
					return

				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Inotify error: %v", err)
			}
		}

	}()

	err = watcher.Add("/direktiv-data")
	if err != nil {
		log.Fatal(err)
	}

	<-done

}

func tail() {

	log.Println("Tailing logs.")

	t, err := tailfile.TailFile(
		"/direktiv-data/out.log", tailfile.Config{Follow: true, ReOpen: true, MustExist: false})
	if err != nil {
		goto end
	}

	for line := range t.Lines {
		log.Printf("Container log: %v", line.Text)
		_, err = flowClient.ActionLog(context.Background(), &flow.ActionLogRequest{
			InstanceId: &instanceId,
			Msg:        []string{line.Text},
		})
		if err != nil {
			log.Printf("Failed to push log: %v.", err)
		}
	}

end:
	log.Println("Log tailing finished.")

}
