package commands

import (
	"encoding/json"
	"fmt"
)

func doubleMarshal[T any](in any) (*T, error) {
	data, ok := in.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("action image configuration has wrong type")
	}

	j, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("action image configuration can not be converted: %s", err.Error())
	}

	var outData T
	if err := json.Unmarshal(j, &outData); err != nil {
		return nil, fmt.Errorf("action image configuration can not be converted: %s", err.Error())
	}

	return &outData, err
}
