package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	instanceSettingsVersion1 = "v1"
)

var ErrInvalidInstanceSettings = errors.New("invalid instance settings")

// InstanceSettings keeps a local copy of various namespace and workflow settings so that the engine doesn't have to look them up separately.
type InstanceSettings struct {
	Version     string // to let us identify and correct outdated versions of this struct
	LogToEvents string
}

func (info *InstanceSettings) MarshalJSON() ([]byte, error) {
	if info == nil {
		return json.Marshal(&instanceSettingsV1{
			Version: instanceSettingsVersion1,
		})
	}

	return json.Marshal(&instanceSettingsV1{
		Version:     instanceSettingsVersion1,
		LogToEvents: info.LogToEvents,
	})
}

type instanceSettingsV1 struct {
	Version         string `json:"version"`
	LogToEvents     string `json:"log_to_events"`
	NamespaceConfig []byte `json:"namespace_config"`
}

func LoadInstanceSettings(data []byte) (*InstanceSettings, error) {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	version, defined := m["version"]
	if !defined {
		return nil, fmt.Errorf("failed to load instance settings: %w: missing version", ErrInvalidInstanceSettings)
	}

	var info *InstanceSettings

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	switch version {
	case instanceSettingsVersion1:
		var v1 instanceSettingsV1
		err = dec.Decode(&v1)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance settings: %w: %w", ErrInvalidInstanceSettings, err)
		}

		info = &InstanceSettings{
			Version:     v1.Version,
			LogToEvents: v1.LogToEvents,
		}

	default:
		return nil, fmt.Errorf("failed to load instance settings: %w: unknown version", ErrInvalidInstanceSettings)
	}

	return info, nil
}
