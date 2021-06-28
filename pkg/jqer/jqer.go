package jqer

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/itchyny/gojq"
)

var (
	StringQueryRequiresWrappings bool
	TrimWhitespaceOnQueryStrings bool
	SearchInStrings              bool
	WrappingBegin                = ""
	WrappingIncrement            = "{{"
	WrappingDecrement            = "}}"
)

/*
	// Existing settings
	StringQueryRequiresWrappings = false
	TrimWhitespaceOnQueryStrings = false
	SearchInStrings              = false
	WrappingBegin                = ""
	WrappingIncrement            = "{{"
	WrappingDecrement            = "}}"
*/

/*
	// New settings
	StringQueryRequiresWrappings = true
	TrimWhitespaceOnQueryStrings = true
	SearchInStrings              = true
	WrappingBegin                = "jq"
	WrappingIncrement            = "("
	WrappingDecrement            = ")"
*/

func Evaluate(data, query interface{}) ([]interface{}, error) {

	if query == nil {
		var out []interface{}
		out = append(out, data)
		return out, nil
	}

	if s, ok := query.(string); ok && !StringQueryRequiresWrappings {
		return jq(data, s)
	}

	return recursiveEvaluate(data, query)

}

func recursiveEvaluate(data, query interface{}) ([]interface{}, error) {

	var out []interface{}

	if query == nil {
		out = append(out, nil)
		return out, nil
	}

	switch query.(type) {
	case bool:
	case int:
	case float64:
	case string:
		return recurseIntoString(data, query.(string))
	case map[string]interface{}:
		return recurseIntoMap(data, query.(map[string]interface{}))
	case []interface{}:
		return recurseIntoArray(data, query.([]interface{}))
	default:
		return nil, fmt.Errorf("unexpected type: %s", reflect.TypeOf(query).String())
	}

	out = append(out, query)

	return out, nil

}

func recurseIntoString(data interface{}, s string) ([]interface{}, error) {

	var out []interface{}
	var offset int

	query := s
	if TrimWhitespaceOnQueryStrings {
		query = strings.TrimSpace(query)
		offset = strings.Index(s, query)
	}

	if !SearchInStrings {
		if strings.HasPrefix(query, WrappingBegin+WrappingIncrement) && strings.HasSuffix(query, WrappingDecrement) {
			query = query[len(WrappingBegin)+len(WrappingIncrement) : len(query)-len(WrappingDecrement)]
			return jq(data, query)
		}
		out = append(out, s)
		return out, nil
	}

	// search in string
	var foundQueries bool
	var stringParts []interface{}
	begin := WrappingBegin + WrappingIncrement

	for {
		idx := strings.Index(query, begin)
		if idx < 0 {
			if len(query) > 0 {
				stringParts = append(stringParts, query)
			}
			break
		}

		if idx > 0 {
			stringParts = append(stringParts, query[:idx])
			offset += idx
			query = query[idx:]
			idx = 0
		}

		counter := 1
		var i int

		for i = len(begin); counter > 0; i++ {

			if i >= len(query) {
				break
			}

			if len(query) >= i+len(WrappingIncrement) {

				c := query[i : i+len(WrappingIncrement)]

				if c == WrappingIncrement {
					counter++
					i += len(WrappingIncrement) - 1
					continue
				}

			}

			if len(query) >= i+len(WrappingDecrement) {

				c := query[i : i+len(WrappingDecrement)]

				if c == WrappingDecrement {
					counter--
					i += len(WrappingDecrement) - 1
					continue
				}

			}

		}

		if counter > 0 {
			return nil, fmt.Errorf("unterminated jq query beginning at offset %v", offset)
		}

		var x []interface{}
		qstr := query[len(begin) : i-1]
		var err error
		foundQueries = true
		x, err = jq(data, qstr)
		if err != nil {
			return nil, fmt.Errorf("error running jq query beginning at offset %v: %v", offset, err)
		}
		out = x

		if len(x) == 0 {
			return nil, fmt.Errorf("error in jq query beginning at offset %v: no results", offset)
		}
		if len(x) > 1 {
			return nil, fmt.Errorf("error in jq query beginning at offset %v: more than one result", offset)
		}

		stringParts = append(stringParts, x[0])

		offset += i
		query = query[i:]

	}

	if !foundQueries {
		out = append(out, s)
		return out, nil
	}

	if len(stringParts) == 1 {
		return out, nil
	}

	var x []string
	for i := range stringParts {
		if _, ok := stringParts[i].(string); ok {
			x = append(x, fmt.Sprintf("%v", stringParts[i]))
		} else {
			data, err := json.Marshal(stringParts[i])
			if err != nil {
				return nil, err
			}
			x = append(x, fmt.Sprintf("%s", data))
		}
	}

	s = strings.Join(x, "")
	out = make([]interface{}, 1)
	out[0] = s
	return out, nil

}

func recurseIntoMap(data interface{}, m map[string]interface{}) ([]interface{}, error) {
	var out []interface{}
	var results = make(map[string]interface{})
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := range keys {
		k := keys[i]
		x, err := recursiveEvaluate(data, m[k])
		if err != nil {
			return nil, fmt.Errorf("error in '%s': %v", k, err)
		}
		if len(x) == 0 {
			return nil, fmt.Errorf("error in element '%s': no results", k)
		}
		if len(x) > 1 {
			return nil, fmt.Errorf("error in element '%s': more than one result", k)
		}
		results[k] = x[0]
	}
	out = append(out, results)
	return out, nil
}

func recurseIntoArray(data interface{}, q []interface{}) ([]interface{}, error) {
	var out []interface{}
	var array = make([]interface{}, 0)
	for i := range q {
		x, err := recursiveEvaluate(data, q[i])
		if err != nil {
			return nil, fmt.Errorf("error in element %d: %v", i, err)
		}
		if len(x) == 0 {
			return nil, fmt.Errorf("error in element %d: no results", i)
		}
		if len(x) > 1 {
			return nil, fmt.Errorf("error in element %d: more than one result", i)
		}
		array = append(array, x[0])
	}
	out = append(out, array)
	return out, nil
}

func jq(input interface{}, command string) ([]interface{}, error) {

	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var x interface{}

	err = json.Unmarshal(data, &x)
	if err != nil {
		return nil, err
	}

	query, err := gojq.Parse(command)
	if err != nil {
		return nil, err
	}

	var output []interface{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	iter := query.RunWithContext(ctx, x)

	for i := 0; ; i++ {

		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return nil, err
		}

		output = append(output, v)

	}

	return output, nil

}
