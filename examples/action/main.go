package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {

	fmt.Printf("Starting!!!!!!\n")

	http.HandleFunc("/", helloServer)
	http.ListenAndServe(":8080", nil)

}

func helloServer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This Is My Data"))
}

func log(aid, l string) {
	http.Post(fmt.Sprintf("http://localhost:8889/log?aid=%s", aid), "plain/text", strings.NewReader(l))
}
