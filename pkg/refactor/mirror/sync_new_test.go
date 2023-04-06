// nolint:testpackage
package mirror

import (
	"reflect"
	"testing"
)

func Test_splitPathToDirectories(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want []string
	}{
		{name: "case", dir: "/a/b/c/d", want: []string{"/a", "/a/b", "/a/b/c", "/a/b/c/d"}},
		{name: "case", dir: "/a", want: []string{"/a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitPathToDirectories(tt.dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitPathToDirectories() = %v, want %v", got, tt.want)
			}
		})
	}
}
