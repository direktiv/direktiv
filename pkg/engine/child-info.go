package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	instanceChildInfoVersion1 = "v1"
)

var ErrInvalidInstanceChildInfo = errors.New("invalid instance child info")

// ChildInfo is part of the ChildrenInfo structure. It represents useful information about a single child action.
type ChildInfo struct {
	ID          string `json:"id"`
	Async       bool   `json:"async"`
	Complete    bool   `json:"complete"`
	Type        string `json:"type"`
	Attempts    int    `json:"attempts"`
	ServiceName string `json:"service_name"`
}

// ChildrenInfo keeps some useful information about all direct child actions of this instance.
type InstanceChildrenInfo struct {
	Version  string // to let us identify and correct outdated versions of this struct
	Children []ChildInfo
}

func (info *InstanceChildrenInfo) MarshalJSON() ([]byte, error) {
	if info == nil {
		return json.Marshal(&InstanceChildInfoV1{
			Version: instanceChildInfoVersion1,
		})
	}

	return json.Marshal(&InstanceChildInfoV1{
		Version:  instanceChildInfoVersion1,
		Children: info.Children,
	})
}

type InstanceChildInfoV1 struct {
	Version  string      `json:"version"`
	Children []ChildInfo `json:"children"`
}

func LoadInstanceChildInfo(data []byte) (*InstanceChildrenInfo, error) {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	version, defined := m["version"]
	if !defined {
		return nil, fmt.Errorf("failed to load instance child info: %w: missing version", ErrInvalidInstanceChildInfo)
	}

	var info *InstanceChildrenInfo

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	switch version {
	case instanceChildInfoVersion1:
		var v1 InstanceChildInfoV1
		err = dec.Decode(&v1)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance child info: %w: %w", ErrInvalidInstanceChildInfo, err)
		}

		info = &InstanceChildrenInfo{
			Version:  v1.Version,
			Children: v1.Children,
		}

	default:
		return nil, fmt.Errorf("failed to load instance child info: %w: unknown version", ErrInvalidInstanceChildInfo)
	}

	return info, nil
}
