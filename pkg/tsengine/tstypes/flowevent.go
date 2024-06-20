package tstypes

import (
	"fmt"

	"github.com/direktiv/direktiv/pkg/utils"
)

type FlowEvent struct {
	Type    string
	Context map[string]interface{}
}

func ParseFlowEvent(v ValueItem) (FlowEvent, error) {
	eventMix, err := unmarshalAndAssert[[]ValueItemMix](v.Value.Value)
	if err != nil {
		return FlowEvent{}, fmt.Errorf("not an array of ValueItemMix for FlowEvent %s: %w", v.Key.Value, err)
	}

	return ParseEvent(eventMix)
}

func ParseEvent(eventMix []ValueItemMix) (FlowEvent, error) {
	event := FlowEvent{}
	for k := range eventMix {
		e := eventMix[k]
		if e.Key.Value == "type" {
			t, err := utils.DoubleMarshal[Key](e.Value)
			if err != nil {
				return event, err
			}
			event.Type = fmt.Sprintf("%v", t.Value)
		}

		if e.Key.Value == "context" {
			vi, err := utils.DoubleMarshal[ValueItem](e)
			if err != nil {
				return event, err
			}
			a := make([]ValueItem, 1)
			a[0] = vi
			err = parseEventValue(a, &event)
			if err != nil {
				return event, err
			}
		}
	}

	return event, nil
}

func ParseFlowEventSlice(v ValueItem) ([]FlowEvent, error) {
	eventsMix, err := unmarshalAndAssert[ValueItemMix](v)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ValueItemMix for []FlowEvent: %w", err)
	}

	ee, err := unmarshalAndAssert[ValueItemList](eventsMix.Value)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ValueItemList for []FlowEvent: %w", err)
	}

	result := make([]FlowEvent, len(ee.Value))
	for i, e := range ee.Value {
		eventMix, err := unmarshalAndAssert[[]ValueItemMix](e.Value)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling ValueItemMix for FlowEvent at index %d: %w", i, err)
		}

		result[i], err = ParseEvent(eventMix)
		if err != nil {
			return nil, fmt.Errorf("error parsing FlowEvent at index %d: %w", i, err)
		}
	}

	return result, nil
}

func parseEventValue(valueList []ValueItem, event *FlowEvent) error {
	// Initialize the context map if it's nil
	if event.Context == nil {
		event.Context = make(map[string]interface{})
	}

	// Check if the number of values to parse matches expected fields for a FlowEvent
	if len(valueList) != 1 {
		return fmt.Errorf("invalid number of values for FlowEvent, expected 1, got %d", len(valueList))
	}

	// Parse context
	context, err := unmarshalAndAssert[map[string]interface{}](valueList[0].Value.Value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal context as map[string]interface{}: %w", err)
	}
	event.Context = context

	return nil
}
