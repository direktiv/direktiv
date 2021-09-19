package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
)

func runAsInit() {

	log.Infof("Running as init container.")

	var lock sync.Mutex
	var received, failed bool
	srv := new(http.Server)

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		lock.Lock()
		if received {
			lock.Unlock()
			log.Infof("Unexpected extra payload request received.")
			code := http.StatusServiceUnavailable
			http.Error(w, http.StatusText(code), code)
			return
		}
		received = true
		lock.Unlock()

		defer func() {
			go srv.Shutdown(context.Background())
		}()

		log.Infof("Initial payload request received.")

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			failed = true
			log.Infof("Error: %v.", err)
			code := http.StatusBadRequest
			http.Error(w, http.StatusText(code), code)
			return
		}

		f, err := os.Create("/direktiv-data/input.json")
		if err != nil {
			failed = true
			log.Infof("Error: %v.", err)
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		_, err = io.Copy(f, bytes.NewReader(data))
		if err != nil {
			failed = true
			log.Infof("Error: %v.", err)
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		err = loadFiles(r)
		if err != nil {
			failed = true
			log.Infof("Error: %v.", err)
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		w.WriteHeader(200)
		_, err = w.Write([]byte("ok"))
		if err != nil {
			log.Infof("Error: %v.", err)
			return
		}

	})

	srv.Addr = "0.0.0.0:8890"
	srv.Handler = router

	var err error
	// k, c, _ := util.CertsForComponent(util.TLSHttpComponent)
	// if len(k) > 0 {
	// 	err = srv.ListenAndServeTLS(c, k)
	// } else {
	err = srv.ListenAndServe()
	// }
	// err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	if failed {
		os.Exit(1)
	}

	log.Infof("Init step completed successfully.")

}
