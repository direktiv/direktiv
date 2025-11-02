package filter

import (
	"net/url"
	"reflect"
	"testing"
)

func TestFromURLValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input url.Values
		want  Values
	}{
		{
			name: "parses eq as default operator",
			input: url.Values{
				"filter[status]": []string{"active"},
			},
			want: Values{
				"status": {OpEq: "active"},
			},
		},
		{
			name: "parses gt and lt",
			input: url.Values{
				"filter[age][gt]": []string{"18"},
				"filter[age][lt]": []string{"30"},
			},
			want: Values{
				"age": {OpGt: "18", OpLt: "30"},
			},
		},
		{
			name: "ignores non-filter params",
			input: url.Values{
				"page":              []string{"2"},
				"filter[price][lt]": []string{"10.0"},
			},
			want: Values{
				"price": {OpLt: "10.0"},
			},
		},
		{
			name: "keeps only first value",
			input: url.Values{
				"filter[tag]": []string{"a", "b"},
			},
			want: Values{
				"tag": {OpEq: "a"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FromURLValues(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("FromURLValues()=%v want %v", got, tt.want)
			}
		})
	}
}

func TestFromQueryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want Values
	}{
		{
			name: "simple eq",
			raw:  "filter[status]=active",
			want: Values{"status": {OpEq: "active"}},
		},
		{
			name: "combined age window",
			raw:  "filter[age][gt]=18&filter[age][lt]=30",
			want: Values{"age": {OpGt: "18", OpLt: "30"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := FromQueryString(tt.raw)
			if err != nil {
				t.Fatalf("FromQueryString(%q) unexpected error: %v", tt.raw, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("FromQueryString(%q)=%v want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestBuildAndFieldHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []func() (string, string, string)
		want Values
	}{
		{
			name: "single eq",
			in:   []func() (string, string, string){FieldEQ("status", "active")},
			want: Values{"status": {OpEq: "active"}},
		},
		{
			name: "gt lt",
			in: []func() (string, string, string){
				FieldGT("age", "18"),
				FieldLT("age", "30"),
			},
			want: Values{"age": {OpGt: "18", OpLt: "30"}},
		},
		{
			name: "mixed fields",
			in: []func() (string, string, string){
				FieldEQ("role", "admin"),
				FieldGT("score", "99.5"),
			},
			want: Values{
				"role":  {OpEq: "admin"},
				"score": {OpGt: "99.5"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Build(tt.in...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Build()=%v want %v", got, tt.want)
			}
		})
	}
}
