package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	instanceRuntimeInfoVersion1 = "v1"
)

var ErrInvalidInstanceRuntimeInfo = errors.New("invalid instance runtime info")

// InstanceRuntimeInfo keeps other miscellaneous information useful to the engine.
type InstanceRuntimeInfo struct {
	Version        string // to let us identify and correct outdated versions of this struct
	Controller     string
	Flow           []string // NOTE: now that we keep a copy of the definition we could replace []string with []int
	StateBeginTime time.Time
	Attempts       int
}

func (info *InstanceRuntimeInfo) MarshalJSON() ([]byte, error) {
	if info == nil {
		return json.Marshal(&instanceRuntimeInfoV1{
			Version: instanceRuntimeInfoVersion1,
		})
	}

	return json.Marshal(&instanceRuntimeInfoV1{
		Version:        instanceRuntimeInfoVersion1,
		Controller:     info.Controller,
		Flow:           info.Flow,
		StateBeginTime: info.StateBeginTime,
		Attempts:       info.Attempts,
	})
}

type instanceRuntimeInfoV1 struct {
	Version        string    `json:"version"`
	Controller     string    `json:"controller"`
	Flow           []string  `json:"flow"`
	StateBeginTime time.Time `json:"state_begin_time"`
	Attempts       int       `json:"attempts"`
}

//nolint:dupl
func LoadInstanceRuntimeInfo(data []byte) (*InstanceRuntimeInfo, error) {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	version, defined := m["version"]
	if !defined {
		return nil, fmt.Errorf("failed to load instance runtime info: %w: missing version", ErrInvalidInstanceRuntimeInfo)
	}

	var info *InstanceRuntimeInfo

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	switch version {
	case instanceRuntimeInfoVersion1:
		var v1 instanceRuntimeInfoV1
		err = dec.Decode(&v1)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance runtime info: %w: %w", ErrInvalidInstanceRuntimeInfo, err)
		}

		info = &InstanceRuntimeInfo{
			Version:        v1.Version,
			Controller:     v1.Controller,
			Flow:           v1.Flow,
			StateBeginTime: v1.StateBeginTime,
			Attempts:       v1.Attempts,
		}

	default:
		return nil, fmt.Errorf("failed to load instance runtime info: %w: unknown version", ErrInvalidInstanceRuntimeInfo)
	}

	return info, nil
}
