package filestore_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

func TestRevisionTags_AddTag(t *testing.T) {
	tests := []struct {
		name string
		tags filestore.RevisionTags
		tag  string
		want filestore.RevisionTags
	}{
		{name: "valid_case", tags: "", tag: "tag1", want: "tag1"},
		{name: "valid_case", tags: ",", tag: "tag1", want: `tag1`},
		{name: "valid_case", tags: ",,", tag: "tag1", want: `tag1`},
		{name: "valid_case", tags: ",tag1", tag: "tag2", want: `tag1,tag2`},
		{name: "valid_case", tags: "tag1,", tag: "tag2", want: `tag1,tag2`},
		{name: "valid_case", tags: "tag1,,", tag: "tag2", want: `tag1,tag2`},
		{name: "valid_case", tags: ",,tag1", tag: "tag2", want: `tag1,tag2`},
		{name: "valid_case", tags: "tag1,", tag: "tag2", want: `tag1,tag2`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tags.AddTag(tt.tag); got != tt.want {
				t.Errorf("AddTag() = >%v<, want >%v<", got, tt.want)
			}
		})
	}
}
