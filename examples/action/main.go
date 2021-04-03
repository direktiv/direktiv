package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type DirektivResponse struct {
	ErrorCode    string      `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	Data         interface{} `json:"data"`
}

type Data struct {
	Jens string
}

func main() {

	fmt.Printf("Starting!!!!!!\n")

	http.HandleFunc("/", helloServer)
	http.ListenAndServe(":8080", nil)

}

func helloServer(w http.ResponseWriter, r *http.Request) {

	h := r.Header.Get("Direktiv-ActionID")

	fmt.Printf("METHIOD %v\n", r.Method)
	fmt.Printf("CL %v\n", r.ContentLength)

	if len(h) > 0 {
		// file, err := os.OpenFile(fmt.Sprintf("/var/log/%s.log", h), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		// if err != nil {
		// 	fmt.Printf("ERR1 %v\n", err)
		// }
		// log.SetOutput(file)
	}

	fmt.Printf("AIDF %v\n", h)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("ERR2 %v\n", err)
	}
	fmt.Printf("Data %v\n", string(data))

	var f DirektivResponse
	var g Data
	g.Jens = string(data)
	f.Data = g
	b, err := json.MarshalIndent(f, "", "\t")
	if err != nil {
		fmt.Printf("ERR3 %v\n", err)
	}

	logGet(h, "This Is MY GET Log")
	logPost(h, "This Is MY POST Log")

	// time.Sleep(20 * time.Second)
	fmt.Fprintf(w, string(b))

}

func logGet(aid, l string) {
	_, err := http.Get(fmt.Sprintf("http://localhost:8889/log?log=%s&aid=%s", l, aid))
	fmt.Printf("DO GET %v", err)
}

func logPost(aid, l string) {
	_, err := http.Post(fmt.Sprintf("http://localhost:8889/log?aid=%s", aid), "plain/text", strings.NewReader(l))
	fmt.Printf("DO POST %v", err)
}
