package utils

import (
	"encoding/json"
	"io"
	"os"
)

const (
	TypeScriptMimeType  = "application/x-typescript"
	TypeScriptExtension = ".direktiv.ts"
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

func CopyFile(src, dst string) (int64, error) {
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	return io.Copy(destination, source)
}
