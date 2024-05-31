package commands

import "encoding/base64"

func Btoa(in string) string {
	return base64.StdEncoding.EncodeToString([]byte(in))
}

func Atob(in string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(in)
	return string(b), err
}
