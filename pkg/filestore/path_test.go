package filestore_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/filestore"
)

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "valid", path: "", want: "/"},
		{name: "valid", path: "/", want: "/"},
		{name: "valid", path: ".", want: "/"},
		{name: "valid", path: "///", want: "/"},
		{name: "valid", path: "/a", want: "/a"},
		{name: "valid", path: "/a/", want: "/a"},
		{name: "valid", path: "/a/b/c/", want: "/a/b/c"},
		{name: "valid", path: "/a//b/c//", want: "/a/b/c"},
		{name: "valid", path: "//a//b/c//", want: "/a/b/c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filestore.SanitizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("SanitizePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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
