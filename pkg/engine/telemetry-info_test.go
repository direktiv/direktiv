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
			want: []byte(`{"version":"v2","traceparent":"","call_path":"","namespace_name":""}`),
		},
		{
			name: "Valid Info",
			info: &InstanceTelemetryInfo{
				Version:       "v2",
				TraceParent:   "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				CallPath:      "/some/path",
				NamespaceName: "namespace1",
			},
			want: []byte(`{"version":"v2","traceparent":"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01","call_path":"/some/path","namespace_name":"namespace1"}`),
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
			name:  "Valid v2 Info",
			input: []byte(`{"version":"v2","traceparent":"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01","call_path":"/some/path","namespace_name":"namespace1"}`),
			want: &InstanceTelemetryInfo{
				Version:       "v2",
				TraceParent:   "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				CallPath:      "/some/path",
				NamespaceName: "namespace1",
			},
		},
		{
			name:  "Valid v2 empty Info",
			input: []byte(`{"version":"v2","traceparent":"","call_path":"","namespace_name":""}`),
			want: &InstanceTelemetryInfo{
				Version:       "v2",
				TraceParent:   "",
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
