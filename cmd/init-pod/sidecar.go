package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
	tailfile "github.com/nxadm/tail"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

type errStruct struct {
	Code    string
	Message string
}

func readErrorFile() (code, msg string, err error) {

	fmt.Println("errfile")
	a, err := ioutil.ReadFile("/direktiv-data/error.json")
	if err != nil {
		return "", "", err
	}
	fmt.Println("errfile1")
	es := new(errStruct)

	err = json.Unmarshal(a, es)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("errfile2 %+v\n", es)
	return es.Code, es.Message, nil

}

func runAsSidecar() {

	log.Infof("Running as sidecar container.")

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

				log.Infof("Inotify event: %v", event)

				if event.Op&fsnotify.Create == fsnotify.Create && event.Name == "/direktiv-data/done" {

					log.Infof("\"Done\" file created.")

					err := setOutVariables(context.Background())
					if err != nil {
						log.Infof("Error setting output file variables: %v", err)
					}

					output, err := ioutil.ReadFile("/direktiv-data/output.json")
					if err != nil {
						log.Infof("Error reading output file: %v", err)
					}

					errCode, errMsg, err := readErrorFile()
					if err != nil {
						log.Infof("Error reading error file: %v", err)
					}

					_, err = flow.ReportActionResults(context.Background(), &grpc.ReportActionResultsRequest{
						InstanceId:   instanceId,
						Step:         step,
						ActionId:     actionId,
						ErrorCode:    errCode,
						ErrorMessage: errMsg,
						Output:       output,
					})
					if err != nil {
						log.Infof("Error reporting action results: %v", err)
						return
					}

					// TODO: file vars
					log.Infof("Job finished.")
					return

				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Infof("Inotify error: %v", err)
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

	log.Infof("Tailing logs.")

	t, err := tailfile.TailFile(
		"/direktiv-data/out.log", tailfile.Config{Follow: true, ReOpen: true, MustExist: false})
	if err != nil {
		goto end
	}

	for line := range t.Lines {
		log.Infof("Container log: %v", line.Text)
		_, err = flow.ActionLog(context.Background(), &grpc.ActionLogRequest{
			InstanceId: instanceId,
			Msg:        []string{line.Text},
		})
		if err != nil {
			log.Infof("Failed to push log: %v.", err)
		}
	}

end:
	log.Infof("Log tailing finished.")

}
