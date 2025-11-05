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

func TestValidatePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     string
		wantOK bool
	}{
		// Valid cases
		{"Root", "/", true},
		{"SimpleAbsolute", "/usr", true},
		{"NestedAbsolute", "/usr/bin", true},
		{"SpecialChars", "/a-b_c.123/x+y@z", true},

		// Invalid: must start with /
		{"Empty", "", false},
		{"NoLeadingSlash", "usr/bin", false},

		// Invalid: trailing slash (except for root)
		{"TrailingSlashSingle", "/usr/", false},
		{"TrailingSlashNested", "/a/b/", false},

		// Invalid: double slashes
		{"DoubleSlashMiddle", "/a//b", false},
		{"DoubleSlashRootOnly", "//", false},
		{"TripleSlashRootPlus", "///a", false},

		// Invalid: dot segments
		{"SingleDot", "/.", false},
		{"DoubleDot", "/..", false},
		{"LeadingDotSegment", "/./a", false},
		{"DotInMiddle", "/a/./b", false},
		{"TrailingDotDot", "/a/..", false},
		{"DotDotInMiddle", "/a/../b", false},

		// Invalid: empty segment
		{"EmptySegmentDueToDoubleSlash", "/a//", false},

		// Invalid: NUL byte
		{"NULByte", "/a\x00b", false},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			got, err := filestore.ValidatePath(tc.in)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("ValidatePath(%q) returned unexpected error: %v", tc.in, err)
				}
				if got != tc.in {
					t.Fatalf("ValidatePath(%q) = %q, want original input", tc.in, got)
				}
			} else {
				if err == nil {
					t.Fatalf("ValidatePath(%q) = %q, want error", tc.in, got)
				}
			}
		})
	}
}
