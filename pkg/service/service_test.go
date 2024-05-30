package service_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/service"
)

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
			wantURL:          "http://foo-bar-file1-05a68ea6d7.kns.svc.cluster.local",
		},
		{
			knativeNamespace: "kns",
			namespace:        "foo",
			typ:              "t1",
			filePath:         "/file1",
			name:             "Bar",
			wantURL:          "http://foo-bar-file1-98db27aa4b.kns.svc.cluster.local",
		},
	}
	for _, tt := range tests {
		t.Run("valid_case", func(t *testing.T) {
			service.SetupGetServiceURLFunc(&core.Config{
				KnativeNamespace: tt.knativeNamespace,
			})

			gotURL := service.GetServiceURL(tt.namespace, tt.typ, tt.filePath, tt.name)
			if gotURL != tt.wantURL {
				t.Errorf("service.GetServiceURL() got: %s, want: %s", gotURL, tt.wantURL)
			}
		})
	}
}
