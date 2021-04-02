package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	h := r.Header.Get("aid")

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

	data, err := io.ReadAll(r.Body)
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
	fmt.Fprintf(w, string(b))

}
