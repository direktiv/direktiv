package utils

import "encoding/json"

const (
	TypeScriptMimeType = "application/x-typescript"
)

func DoubleMarshal[T any](obj interface{}) (T, error) {
	var out T

	in, err := json.Marshal(obj)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(in, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}
