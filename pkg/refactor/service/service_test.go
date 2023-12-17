package service_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/service"
)

func TestGetServiceURL_Docker(t *testing.T) {
	tests := []struct {
		namespace string
		typ       string
		filePath  string
		name      string
		wantURL   string
	}{
		{
			namespace: "foo",
			typ:       "t1",
			filePath:  "/file1",
			name:      "bar",
			wantURL:   "http://foo-bar-file1-ebcc769a15",
		},
	}
	for _, tt := range tests {
		t.Run("valid_case", func(t *testing.T) {
			service.SetupGetServiceURLFunc(&core.Config{}, true)

			gotURL := service.GetServiceURL(tt.namespace, tt.typ, tt.filePath, tt.name)
			if gotURL != tt.wantURL {
				t.Errorf("service.GetServiceURL() got: %s, want: %s", gotURL, tt.wantURL)
			}
		})
	}
}

func TestGetServiceURL_Knative(t *testing.T) {
	tests := []struct {
		knativeNamespace string
		namespace        string
		typ              string
		filePath         string
		name             string
		wantURL          string
	}{
		{
			knativeNamespace: "kns",
			namespace:        "foo",
			typ:              "t1",
			filePath:         "/file1",
			name:             "bar",
			wantURL:          "http://foo-bar-file1-ebcc769a15.kns.svc.cluster.local",
		},
	}
	for _, tt := range tests {
		t.Run("valid_case", func(t *testing.T) {
			service.SetupGetServiceURLFunc(&core.Config{
				KnativeNamespace: tt.knativeNamespace,
			}, false)

			gotURL := service.GetServiceURL(tt.namespace, tt.typ, tt.filePath, tt.name)
			if gotURL != tt.wantURL {
				t.Errorf("service.GetServiceURL() got: %s, want: %s", gotURL, tt.wantURL)
			}
		})
	}
}
