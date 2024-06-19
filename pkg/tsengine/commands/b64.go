package commands

import "encoding/base64"

type BtoaCommand struct{}

func (c *BtoaCommand) GetName() string {
	return "btoa"
}

func (c *BtoaCommand) GetCommandFunction() interface{} {
	return func(in string) string {
		return base64.StdEncoding.EncodeToString([]byte(in))
	}
}

type AtobCommand struct{}

func (c *AtobCommand) GetName() string {
	return "atob"
}

func (c *AtobCommand) GetCommandFunction() interface{} {
	return func(in string) (string, error) {
		b, err := base64.StdEncoding.DecodeString(in)
		return string(b), err
	}
}
