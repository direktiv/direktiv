package flow

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/jqer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func checksum(x interface{}) string {

	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}

	hash, err := computeHash(data)
	if err != nil {
		panic(err)
	}

	return hash

}

func computeHash(data []byte) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func marshal(x interface{}) string {

	data, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(data)

}

func unmarshal(data string, x interface{}) error {

	err := json.Unmarshal([]byte(data), x)
	if err != nil {
		return err
	}

	return nil

}

func unmarshalInstanceInputData(input []byte) interface{} {

	var inputData, stateData interface{}

	err := json.Unmarshal(input, &inputData)
	if err != nil {
		inputData = base64.StdEncoding.EncodeToString(input)
	}

	if _, ok := inputData.(map[string]interface{}); ok {
		stateData = inputData
	} else {
		stateData = map[string]interface{}{
			"input": inputData,
		}
	}

	return stateData

}

func marshalInstanceInputData(input []byte) string {

	x := unmarshalInstanceInputData(input)

	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}

	return string(data)

}

func atobMapBuilder(t reflect.Type, v reflect.Value) interface{} {

	m := make(map[string]interface{})

	iter := v.MapRange()

	for iter.Next() {
		kv := iter.Key()
		vv := iter.Value()

		if !vv.CanInterface() {
			continue
		}

		key := kv.String()

		val := vv.Interface()
		m[key] = val

	}

	x := make(map[string]interface{})

	for k, v := range m {
		x[k] = atobBuilder(v)
	}

	return x

}

func atobStructBuilder(t reflect.Type, v reflect.Value) interface{} {

	m := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {

		if !v.Field(i).CanInterface() {
			continue
		}

		key := t.Field(i).Name

		tag := t.Field(i).Tag.Get("json")
		if tag != "" {
			elems := strings.Split(tag, ",")
			for _, elem := range elems {
				elem = strings.TrimSpace(elem)
				if elem != "omitempty" {
					key = elem
					break
				}
			}
		}

		if key == "-" {
			continue
		}

		val := v.Field(i).Interface()
		m[key] = val

	}

	x := make(map[string]interface{})

	for k, v := range m {
		x[k] = atobBuilder(v)
	}

	return x

}

func atobSliceBuilder(t reflect.Type, v reflect.Value) interface{} {

	s := make([]interface{}, v.Len())

	for i := 0; i < v.Len(); i++ {

		if !v.Index(i).CanInterface() {
			continue
		}

		val := v.Index(i).Interface()
		s[i] = val

	}

	for idx, v := range s {
		s[idx] = atobBuilder(v)
	}

	return s

}

func atobBuilder(a interface{}) interface{} {

	v := reflect.ValueOf(a)

deref:

	t := v.Type()

	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		goto deref
	case reflect.Slice:
		return atobSliceBuilder(t, v)
	case reflect.Struct:
		fallthrough
	case reflect.Map:
		x := v.Interface()
		switch x.(type) {
		case time.Time:
			return timestamppb.New(x.(time.Time))
		case map[string]interface{}:
			return atobMapBuilder(t, v)
		default:
			return atobStructBuilder(t, v)
		}
	default:
		x := v.Interface()
		switch x.(type) {
		default:
			return x
		}
	}

}

func atob(a, b interface{}) error {

	m := atobBuilder(a)

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, b)
	if err != nil {
		return err
	}

	return nil

}

func (srv *server) initJQ() {

	jqer.StringQueryRequiresWrappings = true
	jqer.TrimWhitespaceOnQueryStrings = true

	jqer.SearchInStrings = true
	jqer.WrappingBegin = "jq"
	jqer.WrappingIncrement = "("
	jqer.WrappingDecrement = ")"

}

func jq(input interface{}, command interface{}) ([]interface{}, error) {
	out, err := jqer.Evaluate(input, command)
	if err != nil {
		return nil, NewCatchableError(ErrCodeJQBadQuery, "failed to evaluate jq/js: %v", err)
	}
	return out, nil
}

func jqOne(input interface{}, command interface{}) (interface{}, error) {

	output, err := jq(input, command)
	if err != nil {
		return nil, err
	}

	if len(output) != 1 {
		return nil, NewCatchableError(ErrCodeJQNotObject, "the `jq` or `js` command produced multiple outputs")
	}

	return output[0], nil

}

func jqObject(input interface{}, command interface{}) (map[string]interface{}, error) {

	x, err := jqOne(input, command)
	if err != nil {
		return nil, err
	}

	m, ok := x.(map[string]interface{})
	if !ok {
		return nil, NewCatchableError(ErrCodeJQNotObject, "the `jq` or `js` command produced a non-object output")
	}

	return m, nil

}

func truth(x interface{}) bool {

	var success bool

	if x != nil {
		switch x.(type) {
		case bool:
			if x.(bool) {
				success = true
			}
		case string:
			if x.(string) != "" {
				success = true
			}
		case int:
			if x.(int) != 0 {
				success = true
			}
		case float64:
			if x.(float64) != 0.0 {
				success = true
			}
		case []interface{}:
			if len(x.([]interface{})) > 0 {
				success = true
			}
		case map[string]interface{}:
			if len(x.(map[string]interface{})) > 0 {
				success = true
			}
		default:
		}
	}

	return success

}
