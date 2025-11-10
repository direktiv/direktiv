package filter

import (
	"fmt"
	"net/url"
	"testing"
)

func TestContainsMatchBasic(t *testing.T) {
	tests := []struct {
		name     string
		filter   Values
		field    string
		value    string
		expected bool
	}{
		{
			name:     "needle in middle",
			filter:   Values{"name": {OpContains: "ann"}},
			field:    "name",
			value:    "Joanne",
			expected: true,
		},
		{
			name:     "needle at start",
			filter:   Values{"name": {OpContains: "Jo"}},
			field:    "name",
			value:    "Joanne",
			expected: true,
		},
		{
			name:     "needle at end",
			filter:   Values{"name": {OpContains: "ne"}},
			field:    "name",
			value:    "Joanne",
			expected: true,
		},
		{
			name:     "needle absent",
			filter:   Values{"name": {OpContains: "xyz"}},
			field:    "name",
			value:    "Joanne",
			expected: false,
		},
		{
			name:     "different field unaffected",
			filter:   Values{"title": {OpContains: "ann"}},
			field:    "name",
			value:    "Joanne",
			expected: true, // no filter for 'name' means pass
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.Match(tc.field, tc.value)
			if got != tc.expected {
				t.Fatalf("Match(%q,%q) = %v, want %v", tc.field, tc.value, got, tc.expected)
			}
		})
	}
}

func TestContainsTrimmingAndEmptyNeedle(t *testing.T) {
	tests := []struct {
		name     string
		filter   Values
		value    string
		expected bool
	}{
		{
			name:     "trims both sides",
			filter:   Values{"name": {OpContains: " ann "}},
			value:    "  Joanne  ",
			expected: true,
		},
		{
			name:     "empty needle does not match",
			filter:   Values{"name": {OpContains: ""}},
			value:    "anything",
			expected: false,
		},
		{
			name:     "spaces-only needle does not match",
			filter:   Values{"name": {OpContains: "   "}},
			value:    "anything",
			expected: false,
		},
		{
			name:     "empty haystack does not match",
			filter:   Values{"name": {OpContains: "a"}},
			value:    "",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.filter.Match("name", tc.value)
			if got != tc.expected {
				t.Fatalf("Match(%q) = %v, want %v", tc.value, got, tc.expected)
			}
		})
	}
}

func TestContainsIsCaseSensitive(t *testing.T) {
	f := Values{"name": {OpContains: "Ann"}}
	if got := f.Match("name", "joanne"); got {
		t.Fatalf("expected case-sensitive non-match, got %v", got)
	}
	if got := f.Match("name", "JoAnnette"); !got {
		t.Fatalf("expected case-sensitive match, got %v", got)
	}
}

func TestFromURLValuesParsesContains(t *testing.T) {
	u := url.Values{}
	u.Set(fmt.Sprintf("filter[name][%s]", OpContains), "ann")
	u.Set("filter[irrelevant][eq]", "x")

	got := FromURLValues(u)

	if got["name"][OpContains] != "ann" {
		t.Fatalf("expected contains to parse as 'ann', got %q", got["name"][OpContains])
	}
	if _, ok := got["irrelevant"][OpEq]; !ok {
		t.Fatalf("expected other operators to still parse")
	}
}

func TestWithFieldCONTAINS(t *testing.T) {
	f := With(nil, FieldCONTAINS("title", "Go"))
	if _, ok := f["title"]; !ok {
		t.Fatalf("expected field 'title' to exist")
	}
	if f["title"][OpContains] != "Go" {
		t.Fatalf("expected 'Go', got %q", f["title"][OpContains])
	}

	if !f.Match("title", "Go in Action") {
		t.Fatalf("expected match for 'Go in Action'")
	}
	if f.Match("title", "Rust in Action") {
		t.Fatalf("expected non-match for 'Rust in Action'")
	}
}

func TestContainsWithOtherOperatorsAND(t *testing.T) {
	// name must contain "ann" AND be in the allowed CSV set.
	f := Values{
		"name": {
			OpContains: "ann",
			OpIn:       "ann, Joanne, Hannah",
		},
	}
	tests := []struct {
		val      string
		expected bool
	}{
		{"Joanne", true},  // contains "ann" and in list (string-equal "Joanne")
		{"Hannah", true},  // contains "ann" and in list
		{"Banned", false}, // contains "ann" but NOT in list
		{"ann", true},     // contains "ann" and in list
		{"Joan", false},   // in list? no. contains? no.
	}

	for _, tt := range tests {
		if got := f.Match("name", tt.val); got != tt.expected {
			t.Fatalf("value %q => got %v, want %v", tt.val, got, tt.expected)
		}
	}
}
