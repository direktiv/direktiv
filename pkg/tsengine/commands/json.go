package commands

import (
	"encoding/json"
)

type FromJSONCommand struct{}

func (c *FromJSONCommand) GetName() string {
	return "fromJSON"
}

func (c *FromJSONCommand) GetCommandFunction() interface{} {
	return func(in interface{}) (string, error) {
		b, err := json.Marshal(in)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

type ToJSONCommand struct{}

func (c *ToJSONCommand) GetName() string {
	return "toJSON"
}

func (c *ToJSONCommand) GetCommandFunction() interface{} {
	return func(in string) (interface{}, error) {
		var out interface{}
		err := json.Unmarshal([]byte(in), &out)
		if err != nil {
			return "", err
		}
		return out, nil
	}
}
