package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// ActionResponse is the structure to return from actions
// type ActionResponse struct {
// 	ErrorCode    string      `json:"errorCode"`
// 	ErrorMessage string      `json:"errorMessage"`
// 	Data         interface{} `json:"data"`
// }

func main() {

	fmt.Printf("Starting!!!!!!\n")

	http.HandleFunc("/", helloServer)
	http.ListenAndServe(":8080", nil)

}

func helloServer(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Direktiv-ErrorCode", "com.request.error")

	aid := r.Header.Get("Direktiv-ActionID")
	if len(aid) == 0 {
		w.Header().Add("Direktiv-ErrorMessage", "action id missing")
		return
	}

	defer r.Body.Close()
	in, err := ioutil.ReadAll(r.Body)
	if err != nil {
		txt := fmt.Sprintf("error reading body: %v", err)
		log(aid, txt)
		w.Header().Add("Direktiv-ErrorMessage", txt)
	}

	m := make(map[string]string)
	err = json.Unmarshal(in, &m)
	if err != nil {
		txt := fmt.Sprintf("error reading body: %v", err)
		log(aid, txt)
		w.Header().Add("Direktiv-ErrorMessage", txt)
	}

	resp, err := http.Get(m["url"])
	if err != nil {
		txt := fmt.Sprintf("error get request: %v", err)
		log(aid, txt)
		w.Header().Add("Direktiv-ErrorMessage", txt)
	}

	defer resp.Body.Close()
	in, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		txt := fmt.Sprintf("error get request body: %v", err)
		log(aid, txt)
		w.Header().Add("Direktiv-ErrorMessage", txt)
	}

	if len(w.Header().Get("Direktiv-ErrorMessage")) == 0 {

		if err != nil {
			txt := fmt.Sprintf("error unmarshal: %v %v, %v", err, m["url"], string(in))
			log(aid, txt)
			w.Header().Add("Direktiv-ErrorMessage", txt)
		} else {
			w.Write(in)
			w.Header().Del("Direktiv-ErrorMessage")
			w.Header().Del("Direktiv-ErrorCode")
		}

	}

}

// const (
// 	LvlCrit Lvl = iota
// 	LvlError
// 	LvlWarn
// 	LvlInfo
// 	LvlDebug

func log(aid, l string) {
	http.Post(fmt.Sprintf("http://localhost:8889/log?aid=%s", aid), "plain/text", strings.NewReader(l))
}
