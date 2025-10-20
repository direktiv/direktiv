package filter_test

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/api/filter"
)

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, s)
}
func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	tt, err := parseTime(s)
	if err != nil {
		t.Fatalf("parseTime(%q): %v", s, err)
	}
	return tt
}

func TestParseFilters_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		expect filter.Values
	}{
		{
			name: "basic fields with eq",
			raw:  "filter[path]=foo&filter[status]=failed",
			expect: filter.Values{
				"path":   {"eq": "foo"},
				"status": {"eq": "failed"},
			},
		},
		{
			name: "with gt and lt operators",
			raw:  "filter[createdAt][gt]=2025-09-30T22:00:00Z&filter[createdAt][lt]=2025-10-10T22:00:00Z",
			expect: filter.Values{
				"createdAt": {
					"gt": "2025-09-30T22:00:00Z",
					"lt": "2025-10-10T22:00:00Z",
				},
			},
		},
		{
			name: "mix eq and custom op",
			raw:  "filter[user]=alice&filter[age][gt]=21",
			expect: filter.Values{
				"user": {"eq": "alice"},
				"age":  {"gt": "21"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := filter.FromQueryString(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for field, ops := range tt.expect {
				gotOps, ok := f[field]
				if !ok {
					t.Fatalf("missing field %q in parsed filter", field)
				}
				for op, val := range ops {
					if gotOps[op] != val {
						t.Errorf("field %q[%q] want=%q got=%q", field, op, val, gotOps[op])
					}
				}
			}
		})
	}
}

func TestParseFilters_SkipNonFilterKeys(t *testing.T) {
	values, _ := url.ParseQuery("page=2&filter[status]=failed")
	f := filter.FromURLValues(values)

	if len(f) != 1 {
		t.Fatalf("expected only 1 filter, got %d", len(f))
	}
	if _, ok := f["status"]; !ok {
		t.Fatalf("expected filter 'status' to exist")
	}
}

func TestParse_URLDecoding(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		expect filter.Values
	}{
		// %2F → '/', %40 → '@'
		{
			name: "percent-encoded slash and at-sign",
			raw:  "filter[path]=dir%2Ffoo.wf.ts&filter[user]=alice%40example.com",
			expect: filter.Values{
				"path": {"eq": "dir/foo.wf.ts"},
				"user": {"eq": "alice@example.com"},
			},
		},
		// '+' decodes to space in query strings
		{
			name: "plus decodes to space",
			raw:  "filter[tag]=foo+bar",
			expect: filter.Values{
				"tag": {"eq": "foo bar"},
			},
		},
		// literal '+' must be %2B
		{
			name: "encoded plus preserved as literal plus",
			raw:  "filter[tag]=foo%2Bbar",
			expect: filter.Values{
				"tag": {"eq": "foo+bar"},
			},
		},
		// Unicode ✓ (check mark) — %E2%9C%93
		{
			name: "unicode percent-encoding",
			raw:  "filter[path]=%E2%9C%93",
			expect: filter.Values{
				"path": {"eq": "✓"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filter.FromQueryString(tt.raw)
			if err != nil {
				t.Fatalf("FromQueryString error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("parsed mismatch:\nwant: %#v\n got: %#v", tt.expect, got)
			}
		})
	}
}
