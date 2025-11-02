package filter

import (
	"strconv"
	"testing"
	"time"
)

func TestMatch_EQ(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values Values
		field  string
		input  string
		want   bool
	}{
		{
			name:   "nil values always match",
			values: nil,
			field:  "status",
			input:  "active",
			want:   true,
		},
		{
			name:   "no filter for field -> match",
			values: Values{"other": {OpEq: "x"}},
			field:  "status",
			input:  "active",
			want:   true,
		},
		{
			name:   "eq match true",
			values: Values{"status": {OpEq: "active"}},
			field:  "status",
			input:  "active",
			want:   true,
		},
		{
			name:   "eq mismatch false",
			values: Values{"status": {OpEq: "inactive"}},
			field:  "status",
			input:  "active",
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.values.Match(tt.field, tt.input)
			if got != tt.want {
				t.Fatalf("Match(%q,%q)=%v want %v", tt.field, tt.input, got, tt.want)
			}
		})
	}
}

func TestMatch_GtLt_Int(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values Values
		field  string
		input  string
		want   bool
	}{
		{
			name:   "gt true int",
			values: Values{"age": {OpGt: "21"}},
			field:  "age",
			input:  "22",
			want:   true,
		},
		{
			name:   "gt false int equal",
			values: Values{"age": {OpGt: "21"}},
			field:  "age",
			input:  "21",
			want:   false,
		},
		{
			name:   "lt true int",
			values: Values{"age": {OpLt: "30"}},
			field:  "age",
			input:  "29",
			want:   true,
		},
		{
			name:   "lt false int equal",
			values: Values{"age": {OpLt: "30"}},
			field:  "age",
			input:  "30",
			want:   false,
		},
		{
			name:   "gt and lt ANDed true",
			values: Values{"age": {OpGt: "18", OpLt: "30"}},
			field:  "age",
			input:  "25",
			want:   true,
		},
		{
			name:   "gt and lt ANDed false (too high)",
			values: Values{"age": {OpGt: "18", OpLt: "30"}},
			field:  "age",
			input:  "31",
			want:   false,
		},
		{
			name:   "unknown operator fails safe",
			values: Values{"age": {"weird": "10"}},
			field:  "age",
			input:  "50",
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.values.Match(tt.field, tt.input)
			if got != tt.want {
				t.Fatalf("Match(%q,%q)=%v want %v", tt.field, tt.input, got, tt.want)
			}
		})
	}
}

func TestMatch_GtLt_Float(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values Values
		field  string
		input  string
		want   bool
	}{
		{
			name:   "gt true float",
			values: Values{"price": {OpGt: "19.99"}},
			field:  "price",
			input:  "20.00",
			want:   true,
		},
		{
			name:   "gt false float",
			values: Values{"price": {OpGt: "19.99"}},
			field:  "price",
			input:  "19.50",
			want:   false,
		},
		{
			name:   "lt true float",
			values: Values{"price": {OpLt: "19.99"}},
			field:  "price",
			input:  "19.50",
			want:   true,
		},
		{
			name:   "lt false float",
			values: Values{"price": {OpLt: "19.99"}},
			field:  "price",
			input:  "20.01",
			want:   false,
		},
		{
			name:   "AND float window(true)",
			values: Values{"price": {OpGt: "9.99", OpLt: "10.01"}},
			field:  "price",
			input:  "10.00",
			want:   true,
		},
		{
			name:   "AND float window(false)",
			values: Values{"price": {OpGt: "9.99", OpLt: "10.01"}},
			field:  "price",
			input:  "11.00",
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.values.Match(tt.field, tt.input)
			if got != tt.want {
				t.Fatalf("Match(%q,%q)=%v want %v", tt.field, tt.input, got, tt.want)
			}
		})
	}
}

func TestMatch_GtLt_Time(t *testing.T) {
	t.Parallel()

	// Fixed times for determinism
	baseRFC3339 := "2024-01-01T00:00:00Z"
	laterRFC3339 := "2025-06-15T12:34:56Z"

	// Unix seconds and ms around the same point
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	baseSec := base.Unix() // 1704067200
	later := base.Add(24 * time.Hour)
	laterSec := later.Unix()
	baseMs := base.UnixMilli()
	//laterMs := later.UnixMilli()

	tests := []struct {
		name   string
		values Values
		field  string
		input  string
		want   bool
	}{
		{
			name:   "gt true RFC3339",
			values: Values{"created_at": {OpGt: baseRFC3339}},
			field:  "created_at",
			input:  laterRFC3339,
			want:   true,
		},
		{
			name:   "lt false RFC3339",
			values: Values{"created_at": {OpLt: baseRFC3339}},
			field:  "created_at",
			input:  laterRFC3339,
			want:   false,
		},
		{
			name:   "gt false RFC3339",
			values: Values{"created_at": {OpGt: laterRFC3339}},
			field:  "created_at",
			input:  baseRFC3339,
			want:   false,
		},
		{
			name:   "lt true RFC3339",
			values: Values{"created_at": {OpLt: laterRFC3339}},
			field:  "created_at",
			input:  baseRFC3339,
			want:   true,
		},
		{
			name:   "gt true Unix seconds",
			values: Values{"ts": {OpGt: intToStr(int(baseSec))}},
			field:  "ts",
			input:  intToStr(int(laterSec)),
			want:   true,
		},
		{
			name:   "lt true Unix ms",
			values: Values{"tsms": {OpLt: int64ToStr(baseMs)}},
			field:  "tsms",
			input:  int64ToStr(baseMs - 1),
			want:   true,
		},
		{
			name:   "AND window with RFC3339",
			values: Values{"t": {OpGt: "2024-01-01T00:00:00Z", OpLt: "2024-01-03T00:00:00Z"}},
			field:  "t",
			input:  "2024-01-02T12:00:00Z",
			want:   true,
		},
		{
			name:   "parse failure -> false",
			values: Values{"t": {OpGt: "not-a-time"}},
			field:  "t",
			input:  "2024-01-02T12:00:00Z",
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.values.Match(tt.field, tt.input)
			if got != tt.want {
				t.Fatalf("Match(%q,%q)=%v want %v", tt.field, tt.input, got, tt.want)
			}
		})
	}
}

func intToStr(i int) string     { return fmtInt(int64(i)) }
func int64ToStr(i int64) string { return fmtInt(i) }
func fmtInt(i int64) string     { return strconv.FormatInt(i, 10) }
