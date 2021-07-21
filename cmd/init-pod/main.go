package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	tailfile "github.com/nxadm/tail"
)

func tail() {

	t, err := tailfile.TailFile(
		"/direktiv-data/out.log", tailfile.Config{Follow: true, ReOpen: true, MustExist: false})
	if err != nil {
		panic(err)
	}

	// Print the text of each received line
	for line := range t.Lines {
		fmt.Printf("LOG %v\n", line.Text)
	}

}

func main() {

	if os.Getenv("DIREKTIV_LIFECYCLE") == "init" {

		log.Println("init state")
		http.HandleFunc("/", dataHandler)

		if err := http.ListenAndServe("0.0.0.0:8890", nil); err != nil {
			log.Fatal(err)
		}
	}

	if os.Getenv("DIREKTIV_LIFECYCLE") == "run" {

		log.Println("run state")

		go tail()

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					log.Println("event:", event)
					if event.Op&fsnotify.Create == fsnotify.Create && event.Name == "/direktiv-data/done" {
						log.Println("file done created we can finish now")

						a, err := ioutil.ReadFile("/direktiv-data/error.json")
						if err != nil {
							log.Printf("ERROR FILE %v\n", err)
						}
						log.Printf("ERRORFILE: %v", string(a))

						a, err = ioutil.ReadFile("/direktiv-data/output.json")
						if err != nil {
							log.Printf("OUTPUT FILE %v\n", err)
						}
						log.Printf("OUTPUTFILE: %v", string(a))

						os.Exit(0)

					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()

		err = watcher.Add("/direktiv-data")
		if err != nil {
			log.Fatal(err)
		}
		<-done

		// log.Println("run state. adding routes.")
		// http.HandleFunc("/", runHandler)
		// http.HandleFunc("/healthz", healthHandler)
		// http.HandleFunc("/start", startHandler)
		//
		// if err := http.ListenAndServe("0.0.0.0:8888", nil); err != nil {
		// 	log.Fatal(err)
		// }

	}

}

// func startHandler(w http.ResponseWriter, r *http.Request) {
//
// 	fmt.Println("START HANDLER")
// 	w.WriteHeader(200)
// 	w.Write([]byte("started"))
//
// }
//
// func healthHandler(w http.ResponseWriter, r *http.Request) {
//
// 	fmt.Println("HEALTH HANDLER")
// 	w.WriteHeader(200)
// 	w.Write([]byte("healthy and vegan"))
//
// }

// func runHandler(w http.ResponseWriter, r *http.Request) {
//
// 	fmt.Println("OUTPUT HANDLER")
//
// 	a, err := ioutil.ReadFile("/direktiv-data/error.json")
// 	if err != nil {
// 		log.Printf("ERROR FILE %v\n", err)
// 	}
// 	log.Printf("ERRORFILE: %v", string(a))
//
// 	a, err = ioutil.ReadFile("/direktiv-data/output.json")
// 	if err != nil {
// 		log.Printf("OUTPUT FILE %v\n", err)
// 	}
// 	log.Printf("OUTPUTFILE: %v", string(a))
//
// 	os.Exit(0)
// }

var count = 0

func dataHandler(w http.ResponseWriter, r *http.Request) {

	if count == 0 {
		fmt.Println("GOT DATA IN")

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("ERROR %v\n", err)
		}

		f, err := os.Create("/direktiv-data/input.json")
		if err != nil {
			log.Printf("ERROR %v\n", err)
		}
		f.Write(b)

		fmt.Printf("WRITTEN AS INPUT %v\n", string(b))

		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}

	if count == 1 {
		fmt.Println("again so close")
		os.Exit(0)
	}
	count++

}
