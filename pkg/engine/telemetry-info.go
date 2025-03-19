package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

const (
	telemetryInfoVersion1 = "v1"
)

var ErrInvalidInstanceTelemetryInfo = errors.New("invalid instance telemetry info")

// InstanceTelemetryInfo keeps information useful to our telemetry logic.
type InstanceTelemetryInfo struct {
	Version     string
	TraceParent string
	CallPath    string
}

func (info *InstanceTelemetryInfo) MarshalJSON() ([]byte, error) {
	if info == nil {
		return json.Marshal(&instanceTelemetryInfoV1{
			Version: telemetryInfoVersion1,
		})
	}
	return json.Marshal(&instanceTelemetryInfoV1{
		Version:     telemetryInfoVersion1,
		TraceParent: info.TraceParent,
		CallPath:    info.CallPath,
	})

}

// instanceTelemetryInfoV2 represents the v2 format, where we store traceparent.
type instanceTelemetryInfoV1 struct {
	Version     string `json:"version"`
	TraceParent string `json:"traceparent"`
	CallPath    string `json:"callpath"`
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

	switch version {
	case instanceRuntimeInfoVersion1:
		var v1 instanceTelemetryInfoV1
		err = dec.Decode(&v1)
		if err != nil {
			return nil, fmt.Errorf("failed to load instance telemetry info: %w: %w", ErrInvalidInstanceTelemetryInfo, err)
		}

		info = &InstanceTelemetryInfo{
			Version:     v1.Version,
			TraceParent: v1.TraceParent,
			CallPath:    v1.CallPath,
		}

	default:
		return nil, fmt.Errorf("failed to load instance telemetry info: %w: unknown version", ErrInvalidInstanceRuntimeInfo)
	}

	return info, nil
}

func TraceReconstruct(ctx context.Context, ti *InstanceTelemetryInfo, msg string) (context.Context, trace.Span) {
	// does it have a valid one?
	span := trace.SpanFromContext(ctx)

	// if there is already a trace id add new child
	// same for empty or nil
	if span.SpanContext().HasTraceID() || ti == nil || ti.TraceParent == "" {
		return telemetry.Tracer.Start(ctx, msg)
	}

	// reconstruct
	ctx = telemetry.FromTraceParent(ctx, ti.TraceParent)
	ctx, span = telemetry.Tracer.Start(ctx, msg)

	f := telemetry.TraceParent(ctx)
	ti.TraceParent = f

	return ctx, span
}

func TraceGet(ctx context.Context, ti *InstanceTelemetryInfo) (context.Context, trace.Span) {
	// does it have a valid one?
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return ctx, span
	}

	ctx = telemetry.FromTraceParent(ctx, ti.TraceParent)

	return ctx, trace.SpanFromContext(ctx)
}
