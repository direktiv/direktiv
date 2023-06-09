package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	instanceDescentInfoVersion1 = "v1"
)

var ErrInvalidInstanceDescentInfo = errors.New("invalid instance descent info")

// ParentInfo is part of the InstanceDescentInfo structure. It represents useful information about a single instance in the chain.
type ParentInfo struct {
	ID     uuid.UUID `json:"id"`
	State  string    `json:"state"`
	Step   int       `json:"step"`
	Branch int       `json:"branch"` // NOTE: renamed iterator to branch
}

// InstanceDescentInfo keeps a local copy of useful information about the entire chain of parent instances all the way to the root instance, excepting this instance.
type InstanceDescentInfo struct {
	Version string       // to let us identify and correct outdated versions of this struct
	Descent []ParentInfo // chain of callers from the root instance to the direct parent.
}

func (info *InstanceDescentInfo) MarshalJSON() ([]byte, error) {
	if info == nil {
		return json.Marshal(&InstanceDescentInfoV1{
			Version: instanceDescentInfoVersion1,
		})
	}

	return json.Marshal(&InstanceDescentInfoV1{
		Version: instanceDescentInfoVersion1,
		Descent: info.Descent,
	})
}

type InstanceDescentInfoV1 struct {
	Version string       `json:"version"`
	Descent []ParentInfo `json:"descent"`
}

func LoadInstanceDescentInfo(data []byte) (*InstanceDescentInfo, error) {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	version, defined := m["version"]
	if !defined {
		return nil, fmt.Errorf("failed to load instance descent info: %w: missing version", ErrInvalidInstanceDescentInfo)
	}

	var info *InstanceDescentInfo

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	switch version {
	case instanceDescentInfoVersion1:
		var v1 InstanceDescentInfoV1
		err = dec.Decode(&v1)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance descent info: %w: %w", ErrInvalidInstanceDescentInfo, err)
		}

		info = &InstanceDescentInfo{
			Version: v1.Version,
			Descent: v1.Descent,
		}

	default:
		return nil, fmt.Errorf("failed to load instance descent info: %w: unknown version", ErrInvalidInstanceDescentInfo)
	}

	return info, nil
}
