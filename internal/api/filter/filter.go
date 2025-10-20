package filter

import (
	"net/url"
	"regexp"
)

// Operator constants (you can expand these).
const (
	OpEq = "eq"
	OpGt = "gt"
	OpLt = "lt"
)

// Values represents the parsed structure:
// map[field]map[operator]value.
type Values map[string]map[string]string

func (v Values) Match(field string, value string) bool {
	if v == nil {
		return true
	}
	filterField, ok := v[field]
	if !ok {
		return true
	}

	for op, filterValue := range filterField {
		if op == OpEq {
			if filterValue == value {
				return true
			}
		}
	}

	return false
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
		result[field][op] = vals[0]
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

func Build(items ...func() (string, string, string)) Values {
	res := make(Values)
	for _, item := range items {
		op, field, value := item()
		if _, ok := res[field]; !ok {
			res[field] = make(map[string]string)
		}
		res[field][op] = value
	}

	return res
}

func FieldEQ(filed string, value string) func() (string, string, string) {
	return func() (string, string, string) {
		return filed, OpEq, value
	}
}
