package filestore_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

func TestRevisionTags_AddTag(t *testing.T) {
	tests := []struct {
		name string
		tags string
		tag  string
		want string
	}{
		{name: "case1", tags: "tag1,tag2", tag: "tag", want: "tag1,tag2,tag"},
		{name: "case2", tags: ",tag1,tag2,", tag: "tag", want: "tag1,tag2,tag"},
		{name: "case3", tags: ",tag1,tag2", tag: "tag", want: "tag1,tag2,tag"},
		{name: "case4", tags: "tag1,tag2,", tag: "tag", want: "tag1,tag2,tag"},
		{name: "case5", tags: "tag1,tag2", tag: "tag,,", want: "tag1,tag2,tag"},
		{name: "case6", tags: "tag1,tag2", tag: "tag,tt", want: "tag1,tag2,tag"},
		{name: "case7", tags: "tag1,tag2", tag: ",tag", want: "tag1,tag2,tag"},
		{name: "case8", tags: "tag1,tag2", tag: "tag,", want: "tag1,tag2,tag"},
		{name: "case9", tags: "tag1,tag2", tag: ",tag,", want: "tag1,tag2,tag"},
		{name: "case10", tags: "tag1,tag2", tag: ",tag,", want: "tag1,tag2,tag"},
		{name: "case11", tags: " tag1,tag2 ", tag: ", tag,", want: "tag1,tag2,tag"},
		{name: "case12", tags: "tag1,tag2", tag: ",tag ,", want: "tag1,tag2,tag"},
		{name: "case13", tags: "tag1,tag2", tag: ",tag, tt", want: "tag1,tag2,tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filestore.RevisionTags(tt.tags).AddTag(tt.tag); string(got) != tt.want {
				t.Errorf("AddTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRevisionTags_RemoveTag(t *testing.T) {
	tests := []struct {
		name string
		tags string
		tag  string
		want string
	}{
		{name: "case1", tags: "tag1,tag2", tag: "tag", want: "tag1,tag2"},
		{name: "case2", tags: ",tag1,tag2,", tag: "tag", want: "tag1,tag2"},
		{name: "case3", tags: ",tag1,tag2", tag: "tag", want: "tag1,tag2"},
		{name: "case4", tags: "tag1,tag2,", tag: "tag", want: "tag1,tag2"},
		{name: "case5", tags: "tag1,tag2", tag: "", want: "tag1,tag2"},
		{name: "case6", tags: "tag1,tag2", tag: " ", want: "tag1,tag2"},

		{name: "case7", tags: "tag1,tag2,", tag: "tag1", want: "tag2"},
		{name: "case8", tags: "tag1,tag2,", tag: "tag2", want: "tag1"},

		{name: "case9", tags: "tag1,tag2,", tag: ",tag1", want: "tag2"},
		{name: "case10", tags: "tag1,tag2,", tag: ",tag1,", want: "tag2"},
		{name: "case11", tags: "tag1,tag2,", tag: "tag1,", want: "tag2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filestore.RevisionTags(tt.tags).RemoveTag(tt.tag); string(got) != tt.want {
				t.Errorf("AddTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
