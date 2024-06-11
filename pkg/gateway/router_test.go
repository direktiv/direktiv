package gateway

import (
	"reflect"
	"testing"
)

func TestExtractBetweenCurlyBraces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			"validCase",
			"some {test} with {multiple} {braces}",
			[]string{"test", "multiple", "braces"},
		},
		{
			"validCase",
			"/{test}/{multiple}/{braces}",
			[]string{"test", "multiple", "braces"},
		},
		{
			"validCase",
			"{test}/{multiple}/{braces}/",
			[]string{"test", "multiple", "braces"},
		},
		{
			"validCase",
			"{test/{multiple}/{braces}/",
			[]string{"multiple", "braces"},
		},
		{
			"validCase",
			"some/{braces}/",
			[]string{"braces"},
		},
		{
			"validCase",
			"some/no/param/",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractBetweenCurlyBraces(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractBetweenCurlyBraces() = %v, want %v", got, tt.want)
			}
		})
	}
}
