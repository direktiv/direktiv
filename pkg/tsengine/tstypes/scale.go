package tstypes

import "fmt"

type Scale struct {
	Min    int
	Max    int
	Cron   string
	Metric string
	Value  int
}

func ParseScaleSlice(v ValueItem) ([]Scale, error) {
	arguments, err := unmarshalAndAssert[[]Argument](v.Value.Value)
	if err != nil {
		return nil, fmt.Errorf("not an array of Argument for []Scale: %w", err)
	}

	result := make([]Scale, len(arguments))
	for i, arg := range arguments {
		// Extract the first ValueItem from the slice
		if len(arg.Value) == 0 {
			return nil, fmt.Errorf("missing ValueItem in Argument at index %d", i)
		}
		valueItem := arg.Value[0]

		scale := &Scale{}
		err = parseScaleValue(valueItem, scale) // Pass ValueItem, not []ValueItem
		if err != nil {
			return nil, fmt.Errorf("error parsing Scale at index %d: %w", i, err)
		}
		result[i] = *scale
	}

	return result, nil
}

func parseScaleValue(value ValueItem, scale *Scale) error {
	valueMap, err := unmarshalAndAssert[map[string]interface{}](value.Value)
	if err != nil {
		return fmt.Errorf("scale must be a map[string]interface{}: %w", err)
	}

	for k, v := range valueMap {
		switch k {
		case "min":
			scale.Min, err = unmarshalAndAssert[int](v)
			if err != nil {
				return fmt.Errorf("invalid 'min' value for scale: %w", err)
			}
		case "max":
			scale.Max, err = unmarshalAndAssert[int](v)
			if err != nil {
				return fmt.Errorf("invalid 'max' value for scale: %w", err)
			}
		case "cron":
			scale.Cron, err = unmarshalAndAssert[string](v)
			if err != nil {
				return fmt.Errorf("invalid 'cron' value for scale: %w", err)
			}
		case "metric":
			scale.Metric, err = unmarshalAndAssert[string](v)
			if err != nil {
				return fmt.Errorf("invalid 'metric' value for scale: %w", err)
			}
		case "value":
			scale.Value, err = unmarshalAndAssert[int](v)
			if err != nil {
				return fmt.Errorf("invalid 'value' value for scale: %w", err)
			}
		default:
			return fmt.Errorf("unknown key '%s' in scale", k)
		}
	}

	// Validation (optional)
	if scale.Min > scale.Max {
		return fmt.Errorf("invalid scale: 'min' cannot be greater than 'max'")
	}
	// Add other validations as needed (e.g., cron format)

	return nil
}
