// filter_in_test.go
package filter

import (
	"strconv"
	"testing"
	"time"
)

func TestMatch_In_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		vals  Values
		field string
		input string
		want  bool
	}{
		{
			name:  "member present",
			vals:  Values{"status": {OpIn: "active,inactive,pending"}},
			field: "status",
			input: "inactive",
			want:  true,
		},
		{
			name:  "member absent",
			vals:  Values{"status": {OpIn: "active,inactive,pending"}},
			field: "status",
			input: "deleted",
			want:  false,
		},
		{
			name:  "trims spaces",
			vals:  Values{"tag": {OpIn: "a, b ,  c"}},
			field: "tag",
			input: "b",
			want:  true,
		},
		{
			name:  "empty tokens ignored",
			vals:  Values{"x": {OpIn: ",,a,,"}},
			field: "x",
			input: "a",
			want:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.vals.Match(tt.field, tt.input); got != tt.want {
				t.Fatalf("Match(%q,%q)=%v want %v", tt.field, tt.input, got, tt.want)
			}
		})
	}
}

func TestMatch_In_NumberAndFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		vals  Values
		field string
		input string
		want  bool
	}{
		{
			name:  "int equality by number",
			vals:  Values{"age": {OpIn: "18,21,30"}},
			field: "age",
			input: "21",
			want:  true,
		},
		{
			name:  "int not in set",
			vals:  Values{"age": {OpIn: "18,21,30"}},
			field: "age",
			input: "22",
			want:  false,
		},
		{
			name:  "float equality by value not string form",
			vals:  Values{"price": {OpIn: "19.99, 20, 20.500"}},
			field: "price",
			input: "20.5",
			want:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.vals.Match(tt.field, tt.input); got != tt.want {
				t.Fatalf("Match(%q,%q)=%v want %v", tt.field, tt.input, got, tt.want)
			}
		})
	}
}

func TestMatch_In_Time(t *testing.T) {
	t.Parallel()

	base := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	baseRFC := base.Format(time.RFC3339) // e.g. 2024-06-01T12:00:00Z
	sec := strconv.FormatInt(base.Unix(), 10)
	ms := strconv.FormatInt(base.UnixMilli(), 10)

	tests := []struct {
		name  string
		vals  Values
		field string
		input string
		want  bool
	}{
		{
			name:  "RFC3339 in list",
			vals:  Values{"ts": {OpIn: baseRFC + ",2025-01-01T00:00:00Z"}},
			field: "ts",
			input: baseRFC,
			want:  true,
		},
		{
			name:  "Unix seconds in list",
			vals:  Values{"ts": {OpIn: sec + ",1700000000"}},
			field: "ts",
			input: sec,
			want:  true,
		},
		{
			name:  "Unix ms equality",
			vals:  Values{"ts": {OpIn: ms}},
			field: "ts",
			input: ms,
			want:  true,
		},
		{
			name:  "not in list",
			vals:  Values{"ts": {OpIn: "2024-01-01T00:00:00Z,2024-01-02T00:00:00Z"}},
			field: "ts",
			input: "2024-01-03T00:00:00Z",
			want:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.vals.Match(tt.field, tt.input); got != tt.want {
				t.Fatalf("Match(%q,%q)=%v want %v", tt.field, tt.input, got, tt.want)
			}
		})
	}
}

func TestBuild_FieldIN(t *testing.T) {
	t.Parallel()

	vals := Build(
		FieldEQ("status", "active"),
		FieldIN("role", "admin,editor"),
	)

	if !vals.Match("status", "active") {
		t.Fatalf("expected status eq active to match")
	}
	if !vals.Match("role", "admin") || !vals.Match("role", "editor") {
		t.Fatalf("expected role in {admin,editor} to match")
	}
	if vals.Match("role", "viewer") {
		t.Fatalf("did not expect viewer to match")
	}
}
