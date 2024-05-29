package compiler

import (
	"encoding/json"
	"log/slog"

	"github.com/itchyny/gojq"
)

// runs jq queries on parsed content from goja
// doubleMarshal to get to native types
func jq[T any](query string, obj interface{}) ([]T, error) {
	retVal := make([]T, 0)

	in, err := DoubleMarshal[interface{}](obj)
	if err != nil {
		return retVal, err
	}

	q, err := gojq.Parse(query)
	if err != nil {
		slog.Error("jq query parsing error", slog.String("query", query),
			slog.Any("error", err))
		return retVal, err
	}

	iter := q.Run(in)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			return retVal, err
		}

		o, err := DoubleMarshal[T](v)
		if err != nil {
			slog.Error("jq query double marshal error", slog.String("query", query),
				slog.Any("error", err))
			return retVal, err
		}

		retVal = append(retVal, o)
	}

	return retVal, nil
}

func DoubleMarshal[T any](obj interface{}) (T, error) {
	var out T

	in, err := json.Marshal(obj)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(in, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}
