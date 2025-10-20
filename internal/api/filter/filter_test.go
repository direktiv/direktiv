package filter_test

import (
	"net/url"
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
		expect filter.Filters
	}{
		{
			name: "basic fields with eq",
			raw:  "filter[path]=foo&filter[status]=failed",
			expect: filter.Filters{
				"path":   {"eq": "foo"},
				"status": {"eq": "failed"},
			},
		},
		{
			name: "with gt and lt operators",
			raw:  "filter[createdAt][gt]=2025-09-30T22:00:00Z&filter[createdAt][lt]=2025-10-10T22:00:00Z",
			expect: filter.Filters{
				"createdAt": {
					"gt": "2025-09-30T22:00:00Z",
					"lt": "2025-10-10T22:00:00Z",
				},
			},
		},
		{
			name: "mix eq and custom op",
			raw:  "filter[user]=alice&filter[age][gt]=21",
			expect: filter.Filters{
				"user": {"eq": "alice"},
				"age":  {"gt": "21"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := filter.ParseFiltersFromRaw(tt.raw)
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
	f := filter.ParseFilters(values)

	if len(f) != 1 {
		t.Fatalf("expected only 1 filter, got %d", len(f))
	}
	if _, ok := f["status"]; !ok {
		t.Fatalf("expected filter 'status' to exist")
	}
}
