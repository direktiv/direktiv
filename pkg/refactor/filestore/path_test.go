package filestore_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "valid", path: "", want: "/", wantErr: false},
		{name: "valid", path: "/", want: "/", wantErr: false},
		{name: "valid", path: ".", want: "/", wantErr: false},
		{name: "valid", path: "///", want: "/", wantErr: false},
		{name: "valid", path: "/a", want: "/a", wantErr: false},
		{name: "valid", path: "/a/", want: "/a", wantErr: false},
		{name: "valid", path: "/a/b/c/", want: "/a/b/c", wantErr: false},
		{name: "valid", path: "/a//b/c//", want: "/a/b/c", wantErr: false},
		{name: "valid", path: "//a//b/c//", want: "/a/b/c", wantErr: false},
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
