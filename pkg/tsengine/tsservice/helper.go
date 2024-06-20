package tsservice

import (
	"errors"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/itchyny/gojq"
)

// runs jq queries on parsed content fro goja.
// doubleMarshal to get to native types.
func jq[T any](query string, obj interface{}) ([]T, error) {
	retVal := make([]T, 0)

	in, err := utils.DoubleMarshal[interface{}](obj)
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
			var haltError *gojq.HaltError
			if errors.As(err, &haltError) && haltError.Value() == nil {
				break
			} else if err, ok := v.(error); ok { // Handle other errors if needed
				return retVal, err
			}
		}

		o, err := utils.DoubleMarshal[T](v)
		if err != nil {
			slog.Error("jq query double marshal error", slog.String("query", query),
				slog.Any("error", err))

			return retVal, err
		}

		retVal = append(retVal, o)
	}

	return retVal, nil
}
