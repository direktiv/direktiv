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

// Filters represents the parsed structure:
// map[field]map[operator]value.
type Filters map[string]map[string]string

// ParseFilters parses query parameters like filter[field][op]=value
// and returns a generic Filters map.
func ParseFilters(values url.Values) Filters {
	re := regexp.MustCompile(`^filter\[([^\]]+)\](?:\[(\w+)\])?$`)
	result := make(Filters)

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

// ParseFiltersFromRaw is a helper if you have a raw query string.
func ParseFiltersFromRaw(raw string) (Filters, error) {
	v, err := url.ParseQuery(raw)
	if err != nil {
		return nil, err
	}

	return ParseFilters(v), nil
}
