package engine

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInstanceTelemetryInfo_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		info    *InstanceTelemetryInfo
		want    []byte
		wantErr bool
	}{
		{
			name: "Nil Info",
			info: nil,
			want: []byte(`{"version":"v1","trace_id":"","span_id":"","call_path":"","NamespaceName":""}`),
		},
		{
			name: "Valid Info",
			info: &InstanceTelemetryInfo{
				TraceID:       "trace123",
				SpanID:        "span456",
				CallPath:      "/some/path",
				NamespaceName: "namespace1",
			},
			want: []byte(`{"version":"v1","trace_id":"trace123","span_id":"span456","call_path":"/some/path","NamespaceName":"namespace1"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.info.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !json.Valid(got) {
				t.Errorf("MarshalJSON() returned invalid JSON: %v", string(got))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}

func TestLoadInstanceTelemetryInfo(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  *InstanceTelemetryInfo
	}{
		{
			name:  "Valid v1 Info",
			input: []byte(`{"version":"v1","trace_id":"trace123","span_id":"span456","call_path":"/some/path","NamespaceName":"namespace1"}`),
			want: &InstanceTelemetryInfo{
				Version:       "v1",
				TraceID:       "trace123",
				SpanID:        "span456",
				CallPath:      "/some/path",
				NamespaceName: "namespace1",
			},
		},
		{
			name:  "Valid v1 empty Info",
			input: []byte(`{"version":"v1","trace_id":"","span_id":"","call_path":"","NamespaceName":""}`),
			want: &InstanceTelemetryInfo{
				Version:       "v1",
				TraceID:       "",
				SpanID:        "",
				CallPath:      "",
				NamespaceName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadInstanceTelemetryInfo(tt.input)
			if err != nil {
				t.Errorf("LoadInstanceTelemetryInfo() unexpected error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadInstanceTelemetryInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
