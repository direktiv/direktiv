package filestore_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/filestore"
)

func TestGetPathDepth(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "valid", path: "/a/b/c", want: 3},
		{name: "valid", path: "/a", want: 1},
		{name: "valid", path: "/", want: 0},
		{name: "valid", path: ".", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filestore.GetPathDepth(tt.path)
			if got != tt.want {
				t.Errorf("GetPathDepth() got: %v, want: %v", got, tt.want)

				return
			}
		})
	}
}
