/*
	Validate and sets plain varaibles
*/

package main

import (
	"bytes"
	"os"

	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const code = "com.%s.error"

type isolateInput struct {
	Validate []file     `json:"validate"`
	Setter   []variable `json:"set"`
}

type file struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

// output for the requester container
type output struct {
}

func Log(msg string) {

	fmt.Println(msg)

	if logf == nil {
		return
	}

	fmt.Fprintln(logf, msg)

}

var hasFailed bool

func Error(code, msg string) {

	Log(fmt.Sprintf("ERROR: %s; %s", code, msg))

	if hasFailed {
		return
	}

	hasFailed = true

	m := map[string]string{
		"code": code,
		"msg":  msg,
	}

	data, _ := json.Marshal(m)

	err := ioutil.WriteFile("/direktiv-data/error.json", data, 0755)
	if err != nil {
		Error("", err.Error())
	}

}

func Respond(data []byte) {

	err := ioutil.WriteFile("/direktiv-data/output.json", data, 0755)
	if err != nil {
		Error("", err.Error())
	}

}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Request(input []byte) {

	obj := new(isolateInput)
	err := json.Unmarshal(input, obj)
	if err != nil {
		Error(fmt.Sprintf(code, "unmarshal-input"), err.Error())
		return
	}

	// Validate Variables
	for _, v := range obj.Validate {
		Log(fmt.Sprintf("validating %s", v.Key))
		path := fmt.Sprintf("/direktiv-data/vars/%s", v.Key)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			Error(fmt.Sprintf(code, "bad-path"), err.Error())
			return
		}

		if string(f) != v.Value {
			Log("validation failed")
			Log(fmt.Sprintf("found value = %s", string(f)))
			Log(fmt.Sprintf("given value = %s", v.Value))
			Error(fmt.Sprintf(code, "value-mismatch"), fmt.Sprintf("var %s does not match value inside %s", v.Key, path))
			return
		}
	}

	// Set Variables
	for _, v := range obj.Setter {
		Log(fmt.Sprintf("validating %s", v.Key))
		path := fmt.Sprintf("/direktiv-data/vars/out/%s/%s", v.Scope, v.Key)

		err = ioutil.WriteFile(path, []byte(v.Value), 0644)
		if err != nil {
			Error(fmt.Sprintf(code, "bad-write"), err.Error())
			return
		}
	}

	Respond([]byte{})
}

var logf *os.File

func initialize() error {

	var err error

	logf, err = os.Create("/direktiv-data/out.log")
	if err != nil {
		return err
	}

	return nil

}

func cleanup() {

	if logf != nil {
		logf.Close()
	}

	f, err := os.Create("/direktiv-data/done")
	if err == nil {
		f.Close()
	}

}

func main() {

	err := initialize()
	if err != nil {
		Error("", err.Error())
		return
	}

	defer cleanup()

	input, err := ioutil.ReadFile("/direktiv-data/input.json")
	if err != nil {
		Error("", err.Error())
		return
	}

	Log(fmt.Sprintf("INPUT: %s\n", input))

	Request(input)

}
