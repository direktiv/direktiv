package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	telemetryInfoVersion1 = "v1"
	telemetryInfoVersion2 = "v2"
)

var ErrInvalidInstanceTelemetryInfo = errors.New("invalid instance telemetry info")

// InstanceTelemetryInfo keeps information useful to our telemetry logic.
type InstanceTelemetryInfo struct {
	Version       string // Version of the telemetry info
	TraceParent   string // Used in v2
	CallPath      string
	NamespaceName string
}

// MarshalJSON serializes the struct based on its version (v1 or v2).
func (info *InstanceTelemetryInfo) MarshalJSON() ([]byte, error) {
	if info == nil {
		return json.Marshal(&instanceTelemetryInfoV2{
			Version: telemetryInfoVersion2,
		})
	}

	switch info.Version {
	case telemetryInfoVersion1:
		// For v1, we store TraceID, SpanID, and CallPath
		return json.Marshal(&instanceTelemetryInfoV1{
			Version:       telemetryInfoVersion1,
			CallPath:      info.CallPath,
			NamespaceName: info.NamespaceName,
		})

	default:
		return json.Marshal(&instanceTelemetryInfoV2{
			Version:       telemetryInfoVersion2,
			TraceParent:   info.TraceParent,
			CallPath:      info.CallPath,
			NamespaceName: info.NamespaceName,
		})
	}
}

// Deprecated: instanceTelemetryInfoV1 represents the v1 format of InstanceTelemetryInfo.
type instanceTelemetryInfoV1 struct {
	Version       string `json:"version"`
	TraceID       string `json:"trace_id"`
	SpanID        string `json:"span_id"`
	CallPath      string `json:"call_path"`
	NamespaceName string `json:"namespace_name"`
}

// instanceTelemetryInfoV2 represents the v2 format, where we store traceparent.
type instanceTelemetryInfoV2 struct {
	Version       string `json:"version"`
	TraceParent   string `json:"traceparent"`
	CallPath      string `json:"call_path"`
	NamespaceName string `json:"namespace_name"`
}

// LoadInstanceTelemetryInfo deserializes data and handles different versions (v1 and v2).
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
	// dec.DisallowUnknownFields()

	switch version {
	case telemetryInfoVersion1:
		var v1 instanceTelemetryInfoV1
		err = dec.Decode(&v1)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance telemetry info: %w: %w", ErrInvalidInstanceTelemetryInfo, err)
		}

		info = &InstanceTelemetryInfo{
			Version:       v1.Version,
			CallPath:      v1.CallPath,
			NamespaceName: v1.NamespaceName,
		}

	case telemetryInfoVersion2:
		var v2 instanceTelemetryInfoV2
		err = dec.Decode(&v2)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance telemetry info: %w: %w", ErrInvalidInstanceTelemetryInfo, err)
		}

		info = &InstanceTelemetryInfo{
			Version:       v2.Version,
			TraceParent:   v2.TraceParent,
			CallPath:      v2.CallPath,
			NamespaceName: v2.NamespaceName,
		}

	default:
		return nil, fmt.Errorf("failed to load instance telemetry info: %w: unknown version", ErrInvalidInstanceTelemetryInfo)
	}

	return info, nil
}
