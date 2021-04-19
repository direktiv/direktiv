package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vorteil/direktiv/pkg/direktiv"
)

func main() {

	fmt.Printf("Starting server\n")

	mux := http.NewServeMux()
	mux.HandleFunc("/", helloServer)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		shutDown(srv)
	}()

	srv.ListenAndServe()

}

func shutDown(srv *http.Server) {

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	srv.Shutdown(ctxShutDown)

}

func helloServer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This Is My Data"))
	aid := r.Header.Get(direktiv.DirektivActionIDHeader)
	log(aid, "Hello")

	fmt.Printf("AID %v\n", aid)

	w.Write([]byte("{}"))

}

func log(aid, l string) {
	http.Post(fmt.Sprintf("http://localhost:8889/log?aid=%s", aid), "plain/text", strings.NewReader(l))
}
