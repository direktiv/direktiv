package core

import (
	"testing"
)

func TestServiceFileData_GetID(t *testing.T) {
	tests := []struct {
		service ServiceFileData
		wantID  string
	}{
		{
			service: ServiceFileData{
				Typ:       "type1",
				Name:      "name1",
				Namespace: "ns1",
				FilePath:  "/path/to/file.yaml",
			},
			wantID: "ns1-name1-path-to-file-yaml-91dc229468",
		},
		{
			service: ServiceFileData{
				Typ:       "type1",
				Name:      "Name1",
				Namespace: "ns1",
				FilePath:  "/path/to/file.yaml",
			},
			wantID: "ns1-name1-path-to-file-yaml-7efa875ee1",
		},
	}
	for _, tt := range tests {
		t.Run("valid_case", func(t *testing.T) {
			gotID := tt.service.GetID()
			if gotID != tt.wantID {
				t.Errorf("service.GetID() got: %s, want: %s", gotID, tt.wantID)
			}
		})
	}
}
