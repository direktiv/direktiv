package bytedata

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

	"google.golang.org/protobuf/types/known/timestamppb"
)

func Checksum(x interface{}) string {
	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}

	hash, err := ComputeHash(data)
	if err != nil {
		panic(err)
	}

	return hash
}

func ComputeHash(data []byte) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func Marshal(x interface{}) string {
	data, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(data)
}

func Unmarshal(data string, x interface{}) error {
	err := json.Unmarshal([]byte(data), x)
	if err != nil {
		return err
	}

	return nil
}

func UnmarshalInstanceInputData(input []byte) interface{} {
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

func MarshalInstanceInputData(input []byte) string {
	x := UnmarshalInstanceInputData(input)

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
		switch y := x.(type) {
		case time.Time:
			return timestamppb.New(y)
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

func Atob(a, b interface{}) error {
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
