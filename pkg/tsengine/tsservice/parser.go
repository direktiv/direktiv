package tsservice

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/direktiv/direktiv/pkg/tsengine/tstypes"
	"github.com/dop251/goja/ast"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type jsNameStruct struct {
	Name string `json:"Name"`
	Idx  int    `json:"Idx"`
}

type jsStringValue struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   string `json:"Value"`
}

func ParseDefinitionArgs(dec *json.Decoder) (*tstypes.Definition, error) {
	var arg tstypes.Argument
	if err := dec.Decode(&arg); err != nil {
		return nil, err
	}

	def := tstypes.DefaultDefinition()
	if err := parseValue(arg.Value, def); err != nil {
		return nil, err
	}

	return def, nil
}

func parseCommandArgs(dec *json.Decoder) (*tstypes.Function, error) {
	var argList []tstypes.Argument
	if err := dec.Decode(&argList); err != nil {
		return nil, err
	}

	if len(argList) != 1 {
		return nil, fmt.Errorf("argument length is not 1")
	}

	fn := &tstypes.Function{}
	if err := parseValue(argList[0].Value, fn); err != nil {
		return nil, err
	}

	return fn, nil
}

func parseValue(valueItems []tstypes.ValueItem, item any) error {
	ps := reflect.ValueOf(item) // Get the reflection value of the item

	// Check if the item is a pointer
	if ps.Kind() != reflect.Ptr || ps.IsNil() {
		return fmt.Errorf("item must be a non-nil pointer")
	}

	// Get the element that the pointer is pointing to
	elem := ps.Elem()

	for _, v := range valueItems {
		key := v.Key.Value
		field := elem.FieldByName(cases.Title(language.English, cases.Compact).String(key))

		if !field.IsValid() {
			continue // Skip if the field doesn't exist
		}

		var err error
		var parsedValue any

		switch field.Interface().(type) {
		case bool:
			parsedValue, err = parseBool(v)
		case int:
			parsedValue, err = parseInt(v)
		case string:
			parsedValue, err = parseString(v)
		case []string:
			parsedValue, err = parseStringSlice(v)
		case map[string]interface{}:
			parsedValue, err = parseMapStringInterface(v)
		case map[string]string:
			parsedValue, err = parseMapStringString(v)
		case tstypes.FlowEvent:
			parsedValue, err = tstypes.ParseFlowEvent(v)
		case []tstypes.FlowEvent:
			parsedValue, err = tstypes.ParseFlowEventSlice(v)
		case []tstypes.Scale:
			parsedValue, err = tstypes.ParseScaleSlice(v)
		default:
			return fmt.Errorf("unsupported field type: %T", field.Interface())
		}

		if err != nil {
			return err
		}

		// Check if the field is settable before setting the value
		if field.CanSet() {
			field.Set(reflect.ValueOf(parsedValue))
		}
	}

	return nil
}

func parseBool(v tstypes.ValueItem) (bool, error) {
	boolValue, ok := v.Value.Value.(bool)
	if !ok {
		return false, fmt.Errorf("not bool value for %s", v.Key.Value)
	}

	return boolValue, nil
}

func unmarshalAndAssert[T any](value any) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, fmt.Errorf("error marshalling value: %w", err)
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, fmt.Errorf("error unmarshalling value: %w", err)
	}

	return result, nil
}

func parseInt(v tstypes.ValueItem) (int, error) {
	floatValue, err := unmarshalAndAssert[float64](v.Value.Value)
	if err != nil {
		return 0, fmt.Errorf("not a float value for %s: %w", v.Key.Value, err)
	}

	return int(floatValue), nil
}

func parseString(v tstypes.ValueItem) (string, error) {
	strValue, err := unmarshalAndAssert[string](v.Value.Value)
	if err != nil {
		return "", fmt.Errorf("not a string value for %s: %w", v.Key.Value, err)
	}

	return strValue, nil
}

func parseStringSlice(v tstypes.ValueItem) ([]string, error) {
	jsStrings, err := unmarshalAndAssert[[]jsStringValue](v.Value.Value)
	if err != nil {
		return nil, fmt.Errorf("not an array of jsStringValue for %s: %w", v.Key.Value, err)
	}

	result := make([]string, len(jsStrings))
	for i, sVal := range jsStrings {
		result[i] = sVal.Value
	}

	return result, nil
}

func parseMapStringInterface(v tstypes.ValueItem) (map[string]interface{}, error) {
	valueItems, err := unmarshalAndAssert[[]tstypes.ValueItem](v.Value.Value)
	if err != nil {
		return nil, fmt.Errorf("not an array of ValueItem for %s: %w", v.Key.Value, err)
	}

	result := make(map[string]interface{})
	for _, item := range valueItems {
		var itemValue interface{} // Create a variable to hold the parsed value

		// Recursively parse the value
		err = parseValue([]tstypes.ValueItem{item}, &itemValue) // Pass a pointer to itemValue
		if err != nil {
			return nil, err
		}

		result[item.Key.Value] = itemValue
	}

	return result, nil
}

func parseMapStringString(v tstypes.ValueItem) (map[string]string, error) {
	valueItems, err := unmarshalAndAssert[[]tstypes.ValueItem](v.Value.Value)
	if err != nil {
		return nil, fmt.Errorf("not an array of ValueItem for %s: %w", v.Key.Value, err)
	}

	result := make(map[string]string)
	for _, item := range valueItems {
		strValue, err := parseString(item)
		if err != nil {
			return nil, err
		}
		result[item.Key.Value] = strValue
	}

	return result, nil
}

func validateBodyFunctions(prg *ast.Program) error {
	// everything not being a function declaration is
	// stored in a list
	statements := make([]ast.Statement, 0)
	for i := range prg.Body {
		statement := prg.Body[i]
		_, ok := statement.(*ast.FunctionDeclaration)
		if !ok {
			statements = append(statements, statement)
		}
	}

	s, err := jq[string](`..| .Callee?  | 
		select (.Name != null and .Name != "setupFunction") |
		 .Name`, statements)
	if err != nil {
		return err
	}

	// if there are functions left, that is an error. Only first level functions allowed.
	if len(s) > 0 {
		return fmt.Errorf("function calls in the body are not permitted")
	}

	return nil
}
