package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	telemetryInfoVersion1 = "v1"
)

var ErrInvalidInstanceTelemetryInfo = errors.New("invalid instance telemetry info")

// InstanceTelemetryInfo keeps information useful to our telemetry logic.
type InstanceTelemetryInfo struct {
	Version       string `json:"version"` // to let us identify and correct outdated versions of this struct
	TraceID       string `json:"trace_id"`
	SpanID        string `json:"span_id"`
	CallPath      string `json:"call_path"`
	NamespaceName string `json:"namespace_name"`
}

func (info *InstanceTelemetryInfo) MarshalJSON() ([]byte, error) {
	if info == nil {
		return json.Marshal(&instanceTelemetryInfoV1{
			Version: telemetryInfoVersion1,
		})
	}

	return json.Marshal(&instanceTelemetryInfoV1{
		Version:  telemetryInfoVersion1,
		TraceID:  info.TraceID,
		SpanID:   info.SpanID,
		CallPath: info.CallPath,

		NamespaceName: info.NamespaceName,
	})
}

//nolint:musttag
type instanceTelemetryInfoV1 struct {
	Version  string `json:"version"`
	TraceID  string `json:"trace_id"`
	SpanID   string `json:"span_id"`
	CallPath string `json:"call_path"`

	NamespaceName string
}

//nolint:dupl
func LoadInstanceTelemetryInfo(data []byte) (*InstanceTelemetryInfo, error) {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	version, defined := m["version"]
	if !defined {
		return nil, fmt.Errorf("failed to load instance telemetry info: %w: missing version", ErrInvalidInstanceTelemetryInfo)
	}

	var info *InstanceTelemetryInfo

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	switch version {
	case telemetryInfoVersion1:
		var v1 instanceTelemetryInfoV1
		err = dec.Decode(&v1)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance telemetry info: %w: %w", ErrInvalidInstanceTelemetryInfo, err)
		}

		info = &InstanceTelemetryInfo{
			Version:  v1.Version,
			TraceID:  v1.TraceID,
			SpanID:   v1.SpanID,
			CallPath: v1.CallPath,

			NamespaceName: v1.NamespaceName,
		}

	default:
		return nil, fmt.Errorf("failed to load instance telemetry info: %w: unknown version", ErrInvalidInstanceTelemetryInfo)
	}

	return info, nil
}
