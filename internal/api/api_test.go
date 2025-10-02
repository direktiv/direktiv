package api

import (
	"net/http/httptest"
	"testing"
)

func TestParseQueryParam(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
		key      string
		defStr   string
		defInt   int
		wantStr  string
		wantInt  int
		mode     string // "string" or "int"
	}{
		// ----- string mode -----
		{
			name:     "string present",
			rawQuery: "/?name=alice",
			key:      "name",
			defStr:   "guest",
			wantStr:  "alice",
			mode:     "string",
		},
		{
			name:     "string missing uses default",
			rawQuery: "/?other=abc",
			key:      "name",
			defStr:   "guest",
			wantStr:  "guest",
			mode:     "string",
		},
		{
			name:     "string empty uses default",
			rawQuery: "/?name=",
			key:      "name",
			defStr:   "guest",
			wantStr:  "guest",
			mode:     "string",
		},
		{
			name:     "string multiple picks first",
			rawQuery: "/?name=first&name=second",
			key:      "name",
			defStr:   "guest",
			wantStr:  "first",
			mode:     "string",
		},

		// ----- int mode -----
		{
			name:     "int present valid",
			rawQuery: "/?page=42",
			key:      "page",
			defInt:   1,
			wantInt:  42,
			mode:     "int",
		},
		{
			name:     "int missing uses default",
			rawQuery: "/?other=9",
			key:      "page",
			defInt:   7,
			wantInt:  7,
			mode:     "int",
		},
		{
			name:     "int empty uses default",
			rawQuery: "/?page=",
			key:      "page",
			defInt:   3,
			wantInt:  3,
			mode:     "int",
		},
		{
			name:     "int invalid uses default",
			rawQuery: "/?page=abc",
			key:      "page",
			defInt:   5,
			wantInt:  5,
			mode:     "int",
		},
		{
			name:     "int negative parses",
			rawQuery: "/?page=-8",
			key:      "page",
			defInt:   0,
			wantInt:  -8,
			mode:     "int",
		},
		{
			name:     "int multiple picks first",
			rawQuery: "/?page=7&page=8",
			key:      "page",
			defInt:   0,
			wantInt:  7,
			mode:     "int",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tc.rawQuery, nil)

			switch tc.mode {
			case "string":
				got := ParseQueryParam[string](r, tc.key, tc.defStr)
				if got != tc.wantStr {
					t.Errorf("got %q, want %q", got, tc.wantStr)
				}

			case "int":
				got := ParseQueryParam[int](r, tc.key, tc.defInt)
				if got != tc.wantInt {
					t.Errorf("got %d, want %d", got, tc.wantInt)
				}
			}
		})
	}
}
