package commands

import (
	"encoding/json"
	"fmt"
)

func FromJSON(in interface{}) (string, error) {
	b, err := json.Marshal(in)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(b), nil
}

func ToJSON(in string) (interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(in), &out)
	if err != nil {
		return "", err
	}
	return out, nil
}
