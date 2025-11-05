package jqer

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const data1 = `{
  "userId": 1,
  "id": 1,
  "title": "delectus aut autem",
  "completed": false
}`

func Test001(t *testing.T) {
	var data interface{}
	err := json.Unmarshal([]byte(data1), &data)
	if err != nil {
		panic(err)
	}

	qstr := `"\"Hello, world!\""`

	var query interface{}
	err = json.Unmarshal([]byte(qstr), &query)
	if err != nil {
		panic(err)
	}

	results, err := Evaluate(data, query)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", results)
}

func Test002(t *testing.T) {
	var data interface{}
	err := json.Unmarshal([]byte(data1), &data)
	if err != nil {
		panic(err)
	}

	qstr := `".x = 5"`

	var query interface{}
	err = json.Unmarshal([]byte(qstr), &query)
	if err != nil {
		panic(err)
	}

	results, err := Evaluate(data, query)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", results)
}

func Test003(t *testing.T) {
	var data interface{}
	err := json.Unmarshal([]byte(data1), &data)
	if err != nil {
		panic(err)
	}

	qstr := `"{ x: .completed }"`

	var query interface{}
	err = json.Unmarshal([]byte(qstr), &query)
	if err != nil {
		panic(err)
	}

	results, err := Evaluate(data, query)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", results)
}

func Test004(t *testing.T) {
	StringQueryRequiresWrappings = true
	TrimWhitespaceOnQueryStrings = true
	defer func() {
		StringQueryRequiresWrappings = false
		TrimWhitespaceOnQueryStrings = false
	}()

	var data interface{}
	err := json.Unmarshal([]byte(data1), &data)
	if err != nil {
		panic(err)
	}

	qstr := `"  {{{ x: .completed } }} "`

	var query interface{}
	err = json.Unmarshal([]byte(qstr), &query)
	if err != nil {
		panic(err)
	}

	results, err := Evaluate(data, query)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", results)
}

func Test005(t *testing.T) {
	StringQueryRequiresWrappings = true
	TrimWhitespaceOnQueryStrings = true
	SearchInStrings = true
	WrappingBegin = "jq"
	WrappingIncrement = "("
	WrappingDecrement = ")"
	defer func() {
		StringQueryRequiresWrappings = false
		TrimWhitespaceOnQueryStrings = false
		SearchInStrings = false
		WrappingBegin = ""
		WrappingIncrement = "{{"
		WrappingDecrement = "}}"
	}()

	var data interface{}
	err := json.Unmarshal([]byte(data1), &data)
	if err != nil {
		panic(err)
	}

	qstr := `"Was jq(.id) completed? -- jq(    .completed    )"`

	var query interface{}
	err = json.Unmarshal([]byte(qstr), &query)
	if err != nil {
		panic(err)
	}

	results, err := Evaluate(data, query)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", results)
}

func Test006(t *testing.T) {
	StringQueryRequiresWrappings = true
	TrimWhitespaceOnQueryStrings = true
	SearchInStrings = true
	WrappingBegin = "jq"
	WrappingIncrement = "("
	WrappingDecrement = ")"
	defer func() {
		StringQueryRequiresWrappings = false
		TrimWhitespaceOnQueryStrings = false
		SearchInStrings = false
		WrappingBegin = ""
		WrappingIncrement = "{{"
		WrappingDecrement = "}}"
	}()

	var data interface{}
	err := json.Unmarshal([]byte(data1), &data)
	if err != nil {
		panic(err)
	}

	qstr := `"\"jq(.)\""`

	var query interface{}
	err = json.Unmarshal([]byte(qstr), &query)
	if err != nil {
		panic(err)
	}

	results, err := Evaluate(data, query)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", results)
}

func Test007(t *testing.T) {
	StringQueryRequiresWrappings = true
	TrimWhitespaceOnQueryStrings = true
	SearchInStrings = true
	WrappingBegin = "jq"
	WrappingIncrement = "("
	WrappingDecrement = ")"
	defer func() {
		StringQueryRequiresWrappings = false
		TrimWhitespaceOnQueryStrings = false
		SearchInStrings = false
		WrappingBegin = ""
		WrappingIncrement = "{{"
		WrappingDecrement = "}}"
	}()

	var data interface{}
	err := json.Unmarshal([]byte(data1), &data)
	if err != nil {
		panic(err)
	}

	qstr := `"{\"A\": 5, \"B\": false, \"C\": jq(.), \"D\": \"Hello, jq(.id)\"}"`

	var query interface{}
	err = json.Unmarshal([]byte(qstr), &query)
	if err != nil {
		panic(err)
	}

	results, err := Evaluate(data, query)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", results)
}

type leakInterface interface {
	Y() int
}

type leakStruct struct {
	X int `json:"x"`
}

func (a *leakStruct) Y() int {
	return a.X
}

func TestNoLeak(t *testing.T) {
	test := &leakStruct{X: 5}

	results, err := Evaluate(map[string]interface{}{
		"test": leakInterface(test),
	}, `
	js(
		return Object.keys(data["test"]);         
	)`)
	if err != nil {
		t.Error(err)
		return
	}

	s := fmt.Sprintf("%v", results[0])
	s = strings.TrimSpace(s)

	if s != `["x"]` {
		t.Logf(s, []byte(s))

		t.Fail()
	}
}
