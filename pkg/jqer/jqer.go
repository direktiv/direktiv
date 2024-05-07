package jqer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/bbuck/go-lexer"
	"github.com/dop251/goja"
	"github.com/itchyny/gojq"
)

var (
	StringQueryRequiresWrappings bool
	TrimWhitespaceOnQueryStrings bool

	SearchInStrings   bool
	WrappingBegin     = ""
	WrappingIncrement = "{{"
	WrappingDecrement = "}}"
)

// Evaluate evaluates the data against the query provided and returns the result.
func Evaluate(data, query interface{}) ([]interface{}, error) {
	if query == nil {
		var out []interface{}
		out = append(out, data)

		return out, nil
	}

	x, _ := json.Marshal(data)
	m := make(map[string]interface{})
	_ = json.Unmarshal(x, &m)

	return recursiveEvaluate(m, query)
}

func recursiveEvaluate(data, query interface{}) ([]interface{}, error) {
	var out []interface{}

	if query == nil {
		out = append(out, nil)
		return out, nil
	}

	switch q := query.(type) {
	case bool:
	case int:
	case float64:
	case string:
		return recurseIntoString(data, q)
	case map[string]interface{}:
		return recurseIntoMap(data, q)
	case []interface{}:
		return recurseIntoArray(data, q)
	default:
		return nil, fmt.Errorf("unexpected type: %s", reflect.TypeOf(query).String())
	}

	out = append(out, query)

	return out, nil
}

const (
	JqStartToken lexer.TokenType = iota
	JsStartToken
	StringToken
	ErrorToken
	NoToken
)

func JqState(l *lexer.L) lexer.StateFunc {
	src := make([]string, 3)
	var jdxJ int

	mover := func(rewind int, forward bool) {
		//nolint:intrange
		for a := 0; a < rewind; a++ {
			if forward {
				l.Next()
			} else {
				l.Rewind()
			}
		}
	}

	//nolint:intrange
	for i := 0; i < 3; i++ {
		r := l.Next()
		if r == lexer.EOFRune {
			// emit string token if there is content in it
			if len(l.Current()) > 0 {
				l.Emit(StringToken)
			}

			return nil
		}
		src[i] = string(r)

		// if one of the strings has a j we store the index for rewind
		// this is only to save scanning
		if src[i] == "j" && i > 0 {
			jdxJ = i
		}
	}

	isJX := strings.Join(src, "")

	token := NoToken
	if isJX == "jq(" {
		token = JqStartToken
	} else if isJX == "js(" {
		token = JsStartToken
	}

	if token != NoToken {
		// this cuts out the 'jX(' bit
		mover(3, false)

		// emit string token if there is content in it
		if len(l.Current()) > 0 {
			l.Emit(StringToken)
		}
		mover(3, true)

		// counting the '()'
		var open int
		l.Ignore()
		for {
			n := l.Next()
			if n == lexer.EOFRune {
				l.Emit(ErrorToken)
				return nil
			}

			switch n {
			case '(':
				open++
			case ')':
				open--
			}

			if open < 0 {
				l.Rewind()
				break
			}
		}
		l.Emit(token)

		// remove closing ')'
		mover(1, true)
		l.Ignore()

		return JqState
	}

	// only rewind to jdxJ, if there was no j in the runes, we can skip rewind all together
	if jdxJ > 0 {
		mover(3-jdxJ, false)
	}

	return JqState
}

func recurseIntoString(data interface{}, s string) ([]interface{}, error) {
	out := make([]interface{}, 0)

	if TrimWhitespaceOnQueryStrings {
		s = strings.TrimSpace(s)
	}

	l := lexer.New(s, JqState)
	l.Start()

	for {
		tok, done := l.NextToken()
		if done {
			break
		}

		switch tok.Type {
		case ErrorToken:
			return nil, fmt.Errorf("jq/js script missing bracket")
		case JqStartToken:
			x, err := jq(data, tok.Value)
			if err != nil {
				return nil, fmt.Errorf("error executing jq query %s: %w", tok.Value, err)
			}

			if len(x) == 0 || len(x) > 0 && x[0] == nil {
				return nil, fmt.Errorf("error in jq query %s: no results", tok.Value)
			}

			if len(x) == 1 {
				out = append(out, x[0])
			} else {
				return nil, fmt.Errorf("jq query produced multiple outputs")
			}

		case JsStartToken:

			vm := goja.New()

			fn := fmt.Sprintf("function fn(data) {\n %s \n}", tok.Value)
			_, err := vm.RunString(fn)
			if err != nil {
				return nil, fmt.Errorf("error loading js query %s: %w", tok.Value, err)
			}

			fnExe, ok := goja.AssertFunction(vm.Get("fn"))
			if !ok {
				return nil, fmt.Errorf("error getting js query %s: %w", tok.Value, err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			done := make(chan bool, 1)

			go func(ctx context.Context, rt *goja.Runtime, b chan bool) {
				select {
				case <-b:
					return
				case <-ctx.Done():
					rt.Interrupt("timeout")
				}
			}(ctx, vm, done)

			defer func(b chan bool) {
				b <- true
			}(done)

			// decoding base64
			// nolint:errcheck
			vm.Set("atob", func(txt string) string {
				r, err := base64.StdEncoding.DecodeString(txt)
				if err != nil {
					return err.Error()
				}

				return string(r)
			})

			// encoding base64
			// nolint:errcheck
			vm.Set("btoa", func(txt string) string {
				return base64.StdEncoding.EncodeToString([]byte(txt))
			})

			// execute and get results
			v, err := fnExe(goja.Undefined(), vm.ToValue(data))
			if err != nil {
				return nil, fmt.Errorf("error running js query %s: %w", tok.Value, err)
			}

			ret := v.Export()
			if ret == nil {
				return nil, fmt.Errorf("error in js query %s: no results", tok.Value)
			}
			out = append(out, ret)

		default:
			out = append(out, tok.Value)
		}
	}

	if len(out) == 1 {
		return out, nil
	}

	x := make([]string, len(out))
	for i := range out {
		part := out[i]
		if _, ok := part.(string); ok {
			x = append(x, fmt.Sprintf("%v", part))
		} else {
			data, err := json.Marshal(part)
			if err != nil {
				return nil, err
			}
			x = append(x, string(data))
		}
	}

	s = strings.Join(x, "")
	out = make([]interface{}, 1)
	out[0] = s

	return out, nil
}

func recurseIntoMap(data interface{}, m map[string]interface{}) ([]interface{}, error) {
	var out []interface{}
	results := make(map[string]interface{})
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := range keys {
		k := keys[i]
		x, err := recursiveEvaluate(data, m[k])
		if err != nil {
			return nil, fmt.Errorf("error in '%s': %w", k, err)
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
	array := make([]interface{}, 0)
	for i := range q {
		x, err := recursiveEvaluate(data, q[i])
		if err != nil {
			return nil, fmt.Errorf("error in element %d: %w", i, err)
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
