package tstypes

import (
	"encoding/json"
	"fmt"
)

type Argument struct {
	LeftBrace  int         `json:"LeftBrace"`
	RightBrace int         `json:"RightBrace"`
	Value      []ValueItem `json:"Value"`
}

type ValueItemList struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   []struct {
		LeftBrace  int `json:"LeftBrace"`
		RightBrace int `json:"RightBrace"`
		Value      []struct {
			Computed bool        `json:"Computed"`
			Key      Key         `json:"Key"`
			Kind     string      `json:"Kind"`
			Value    interface{} `json:"Value,omitempty"`
		} `json:"Value"`
	} `json:"Value"`
}

type ValueItemMix struct {
	Computed bool   `json:"Computed"`
	Key      Key    `json:"Key"`
	Kind     string `json:"Kind"`
	Value    interface{}
}

type ValueItem struct {
	Computed bool   `json:"Computed"`
	Key      Key    `json:"Key"`
	Kind     string `json:"Kind"`
	Value    struct {
		Idx     int         `json:"Idx"`
		Literal string      `json:"Literal"`
		Value   interface{} `json:"Value"`
	}
}

type Key struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   string `json:"Value"`
}

type FlowInformation struct {
	Definition *Definition
	Messages   *Messages

	Functions map[string]Function
	ID        string
}

func unmarshalAndAssert[T any](value any) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, fmt.Errorf("error marshalling value: %w", err)
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, fmt.Errorf("error unmarshalling value: %w", err)
	}

	return result, nil
}
