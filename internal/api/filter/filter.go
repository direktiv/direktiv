package filter

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Operator constants.
const (
	OpEq = "eq"
	OpGt = "gt"
	OpLt = "lt"
	OpIn = "in"
)

// Values represents the parsed structure:
// map[field]map[operator]value.
type Values map[string]map[string]string

// Match returns true if the given (field,value) satisfies ALL filter operators
// configured for that field. If no filter is present for the field, it returns true.
func (v Values) Match(field string, value string) bool {
	if v == nil {
		return true
	}
	filterField, ok := v[field]
	if !ok {
		return true
	}

	for op, filterValue := range filterField {
		switch op {
		case OpEq:
			if filterValue != value {
				return false
			}
		case OpGt, OpLt:
			// Compare the candidate against the threshold (filter value).
			cmp, ok := compareScalars(value, filterValue)
			if !ok {
				// We couldn't parse both sides as number/float/time -> fail the match for safety.
				return false
			}
			if op == OpGt && (cmp <= 0) {
				return false
			}
			if op == OpLt && (cmp >= 0) {
				return false
			}
		case OpIn:
			if !containsIn(value, filterValue) {
				return false
			}
		default:
			// TODO: Check here for unknown operator.
			return false
		}
	}

	return true
}

// containsIn checks whether 'val' is contained in the CSV 'csvList'.
// Membership uses numeric/time-aware equality via compareScalars (cmp==0) OR
// exact string equality (after trimming). Empty tokens are ignored.
func containsIn(val, csvList string) bool {
	val = strings.TrimSpace(val)
	csvList = strings.TrimSpace(csvList)
	if val == "" || csvList == "" {
		return false
	}

	items := splitCSV(csvList)
	for _, it := range items {
		if it == "" {
			continue
		}
		// First, try scalar-aware equality.
		if cmp, ok := compareScalars(val, it); ok && cmp == 0 {
			return true
		}
		// Fallback to plain string equality (useful for non-numeric/time strings).
		if val == it {
			return true
		}
	}

	return false
}

// splitCSV splits on commas, trims spaces, and returns tokens (may include empty if consecutive commas).
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}

// compareScalars tries (in order) int, float, time.
// Returns negative if a<b, zero if a==b, positive if a>b, and ok=false if no type matched.
func compareScalars(a, b string) (cmp int, ok bool) {
	// Try integers
	if ai, aErr := strconv.ParseInt(strings.TrimSpace(a), 10, 64); aErr == nil {
		if bi, bErr := strconv.ParseInt(strings.TrimSpace(b), 10, 64); bErr == nil {
			switch {
			case ai < bi:
				return -1, true
			case ai > bi:
				return 1, true
			default:
				return 0, true
			}
		}
	}

	// Try floats
	if af, aErr := strconv.ParseFloat(strings.TrimSpace(a), 64); aErr == nil {
		if bf, bErr := strconv.ParseFloat(strings.TrimSpace(b), 64); bErr == nil {
			switch {
			case af < bf:
				return -1, true
			case af > bf:
				return 1, true
			default:
				return 0, true
			}
		}
	}

	// Try times
	at, aok := parseTime(a)
	bt, bok := parseTime(b)
	if aok && bok {
		if at.Before(bt) {
			return -1, true
		}
		if at.After(bt) {
			return 1, true
		}

		return 0, true
	}

	return 0, false
}

// parseTime supports common layouts plus Unix seconds and milliseconds.
func parseTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}

	// Unix ms / s detection
	if isAllDigits(s) {
		// 13 digits -> milliseconds, 10 digits -> seconds.
		if len(s) == 13 {
			if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
				sec := ms / 1000
				nsec := (ms % 1000) * int64(time.Millisecond)

				return time.Unix(sec, nsec).UTC(), true
			}
		}
		if len(s) == 10 {
			if sec, err := strconv.ParseInt(s, 10, 64); err == nil {
				return time.Unix(sec, 0).UTC(), true
			}
		}
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}

	return time.Time{}, false
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return s != ""
}

// FromURLValues parses query parameters like filter[field][op]=value
// and returns a Values.
func FromURLValues(values url.Values) Values {
	re := regexp.MustCompile(`^filter\[([^\]]+)\](?:\[(\w+)\])?$`)
	result := make(Values)

	for key, vals := range values {
		matches := re.FindStringSubmatch(key)
		if len(matches) == 0 {
			continue // skip non-filter params
		}

		field := matches[1]
		op := matches[2]
		if op == "" {
			op = OpEq // default operator
		}

		if _, ok := result[field]; !ok {
			result[field] = make(map[string]string)
		}
		if len(vals) > 0 {
			result[field][op] = vals[0]
		}
	}

	return result
}

// FromQueryString is a helper if you have a raw query string.
func FromQueryString(raw string) (Values, error) {
	v, err := url.ParseQuery(raw)
	if err != nil {
		return nil, err
	}

	return FromURLValues(v), nil
}

func With(base Values, items ...func() (string, string, string)) Values {
	dest := cloneValues(base)
	for _, item := range items {
		op, field, value := item()
		if _, ok := dest[field]; !ok {
			dest[field] = make(map[string]string)
		}
		dest[field][op] = value
	}

	return dest
}

func FieldEQ(field string, value string) func() (string, string, string) {
	return func() (string, string, string) {
		return OpEq, field, value
	}
}

func FieldGT(field string, value string) func() (string, string, string) {
	return func() (string, string, string) {
		return OpGt, field, value
	}
}

func FieldLT(field string, value string) func() (string, string, string) {
	return func() (string, string, string) {
		return OpLt, field, value
	}
}

func FieldIN(field string, csv string) func() (string, string, string) {
	return func() (string, string, string) {
		return OpIn, field, csv
	}
}

func cloneValues(m Values) Values {
	clone := make(Values, len(m))
	for key, innerMap := range m {
		innerClone := make(map[string]string, len(innerMap))
		for k, v := range innerMap {
			innerClone[k] = v
		}
		clone[key] = innerClone
	}

	return clone
}
