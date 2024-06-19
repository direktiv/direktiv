package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"slices"

	"dario.cat/mergo"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/dop251/goja/ast"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Argument struct {
	LeftBrace  int         `json:"LeftBrace"`
	RightBrace int         `json:"RightBrace"`
	Value      []ValueItem `json:"Value"`
}

type ArgumentMix struct {
	LeftBrace  int            `json:"LeftBrace"`
	RightBrace int            `json:"RightBrace"`
	Value      []ValueItemMix `json:"Value"`
}

type ValueItemList struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   []struct {
		LeftBrace  int `json:"LeftBrace"`
		RightBrace int `json:"RightBrace"`
		Value      []struct {
			Computed bool        `json:"Computed"`
			Key      Key         `json:"Key"`
			Kind     string      `json:"Kind"`
			Value    interface{} `json:"Value,omitempty"`
		} `json:"Value"`
	} `json:"Value"`
}

type ValueItemMix struct {
	Computed bool   `json:"Computed"`
	Key      Key    `json:"Key"`
	Kind     string `json:"Kind"`
	Value    interface{}
}

type ValueItem struct {
	Computed bool   `json:"Computed"`
	Key      Key    `json:"Key"`
	Kind     string `json:"Kind"`
	Value    struct {
		Idx     int         `json:"Idx"`
		Literal string      `json:"Literal"`
		Value   interface{} `json:"Value"`
	}
}

type Key struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   string `json:"Value"`
}

type Messages struct {
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

func newMessages() *Messages {
	return &Messages{
		Warnings: make([]string, 0),
		Errors:   make([]string, 0),
	}
}

func (m *Messages) addError(format string, args ...any) {
	m.Errors = append(m.Errors, fmt.Sprintf(format, args...))
}

func (m *Messages) addWarning(format string, args ...any) {
	m.Warnings = append(m.Warnings, fmt.Sprintf(format, args...))
}

func (m *Messages) merge(a *Messages) {
	m.Errors = slices.Concat(m.Errors, a.Errors)
	m.Warnings = slices.Concat(m.Warnings, a.Warnings)
}

type FlowInformation struct {
	Definition *Definition
	messages   *Messages

	Functions map[string]Function
	Secrets   []Secret
	Files     []File

	ID string
}

type jsNameStruct struct {
	Name string `json:"Name"`
	Idx  int    `json:"Idx"`
}

type jsStringValue struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   string `json:"Value"`
}

func (c *Compiler) getID() string {

	str := fmt.Sprintf("%s-%s", "TODO: NAMESPACE", c.JavaScript)
	sh := sha256.Sum256([]byte(str))

	whitelist := regexp.MustCompile("[^a-zA-Z0-9]+")
	str = whitelist.ReplaceAllString(str, "-")

	// Prevent too long ids
	if len(str) > 50 {
		str = str[:50]
	}

	return fmt.Sprintf("%s-%x", str, sh[:5])
}

func (c *Compiler) CompileFlow() (*FlowInformation, error) {

	fileHolder := make(map[string]File)
	secretsHolder := make(map[string]Secret)

	flowInformation := &FlowInformation{
		Definition: DefaultDefinition(),
		messages:   newMessages(),
		Functions:  make(map[string]Function),
		Secrets:    make([]Secret, 0),
		ID:         c.getID(),
		// Files:      make(map[string]File),
		Files: make([]File, 0),
	}

	astIn, err := json.MarshalIndent(c.ast.Body, "", "   ")
	if err != nil {
		return flowInformation, err
	}

	targetName := ""
	calleeName := ""

	dec := json.NewDecoder(bytes.NewReader(astIn))
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}

		if t == "Target" {
			var d jsNameStruct
			err = dec.Decode(&d)
			if err != nil {
				return flowInformation, err
			}
			targetName = d.Name
		}

		if t == "Callee" {
			var d jsNameStruct
			err = dec.Decode(&d)
			if err != nil {
				return flowInformation, err
			}
			calleeName = d.Name
		}

		if t == "ArgumentList" && calleeName == "getFile" {
			f := &File{}
			err = parseCommandArgs(dec, f)
			if err != nil {
				return flowInformation, err
			}

			defaultFile := DefaultFile()
			err := mergo.Merge(f, defaultFile)
			if err != nil {
				return flowInformation, err
			}

			flowInformation.messages.merge(f.Validate())
			fileHolder[(*f).Base64()] = *f
		}

		if t == "ArgumentList" && calleeName == "getSecret" {
			s := &Secret{}
			err = parseCommandArgs(dec, s)
			if err != nil {
				return flowInformation, err
			}
			flowInformation.messages.merge(s.Validate())
			secretsHolder[(*s).Base64()] = *s
		}

		if t == "ArgumentList" && calleeName == "setupFunction" {
			f := &Function{}
			err = parseCommandArgs(dec, f)
			if err != nil {
				return flowInformation, err
			}
			flowInformation.messages.merge(f.Validate())
			flowInformation.Functions[f.GetID()] = *f
		}

		if t == "Initializer" && targetName == "flow" {
			def := &Definition{}
			err = parseDefinitionArgs(dec, def)
			if err != nil {
				return flowInformation, err
			}

			err := mergo.Merge(flowInformation.Definition, def)
			if err != nil {
				return flowInformation, err
			}
			flowInformation.messages.merge(def.Validate())
			flowInformation.Definition = def
		}
	}

	// if no state, we pick the first function
	if flowInformation.Definition.State == "" {
		for i := range c.ast.Body {
			statement := c.ast.Body[i]
			a, ok := statement.(*ast.FunctionDeclaration)
			if ok {
				flowInformation.Definition.State = a.Function.Name.Name.String()
				break
			}
		}
	}

	for _, v := range fileHolder {
		flowInformation.Files = append(flowInformation.Files, v)
	}

	for _, v := range secretsHolder {
		flowInformation.Secrets = append(flowInformation.Secrets, v)
	}

	return flowInformation, nil
}

func parseDefinitionArgs(dec *json.Decoder, item *Definition) error {
	var arg Argument
	err := dec.Decode(&arg)
	if err != nil {
		return err
	}

	a := make([]Argument, 1)
	a[0] = arg

	return parseValue(arg.Value, item)
}

func parseCommandArgs(dec *json.Decoder, item any) error {
	var argList []Argument
	err := dec.Decode(&argList)
	if err != nil {
		return err
	}

	if len(argList) != 1 {
		return fmt.Errorf("argument length is not 1")
	}

	return parseValue(argList[0].Value, item)
}

func parseValue(valueItems []ValueItem, item any) error {

	ps := reflect.ValueOf(item)

	for a := range valueItems {
		v := valueItems[a]
		key := v.Key.Value
		f := ps.Elem().FieldByName(cases.Title(language.English, cases.Compact).String(key))

		switchType := f.Interface()
		switch switchType.(type) {
		case bool:
			boolValue, ok := v.Value.Value.(bool)
			if !ok {
				return fmt.Errorf("not bool value for %s", key)
			}
			f.SetBool(boolValue)
		case int:
			intValue, ok := v.Value.Value.(float64)
			if !ok {
				return fmt.Errorf("not int value for %s", key)
			}
			f.SetInt(int64(intValue))
		case string:
			f.SetString(fmt.Sprintf("%v", v.Value.Value))
		case []string:
			jsStrings, err := utils.DoubleMarshal[[]jsStringValue](v.Value.Value)
			if err != nil {
				return err
			}

			m := reflect.MakeSlice(reflect.TypeOf(make([]string, 0)), len(jsStrings), len(jsStrings))
			for k := range jsStrings {
				sVal := jsStrings[k]
				m.Index(k).SetString(sVal.Value)
			}
			f.Set(m)
		case map[string]interface{}:
			valueItems, err := utils.DoubleMarshal[[]ValueItem](v.Value.Value)
			if err != nil {
				return err
			}

			var i interface{}
			mapOf := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(&i).Elem())
			m := reflect.MakeMapWithSize(mapOf, len(valueItems))
			for k := range valueItems {
				m.SetMapIndex(reflect.ValueOf(valueItems[k].Key.Value),
					reflect.ValueOf(valueItems[k].Value.Value))
			}
			f.Set(m)
		case map[string]string:
			valueItems, err := utils.DoubleMarshal[[]ValueItem](v.Value.Value)
			if err != nil {
				return err
			}

			mapOf := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(""))
			m := reflect.MakeMapWithSize(mapOf, len(valueItems))
			for k := range valueItems {
				m.SetMapIndex(reflect.ValueOf(valueItems[k].Key.Value),
					reflect.ValueOf(valueItems[k].Value.Value))
			}
			f.Set(m)
		case FlowEvent:
			eventMix, err := utils.DoubleMarshal[[]ValueItemMix](v.Value.Value)
			if err != nil {
				return err
			}

			event, err := parseEvent(eventMix)
			if err != nil {
				return err
			}

			f.Set(reflect.ValueOf(event))
		case []FlowEvent:
			eventsMix, err := utils.DoubleMarshal[ValueItemMix](v)
			if err != nil {
				return err
			}

			ee, err := utils.DoubleMarshal[ValueItemList](eventsMix.Value)
			if err != nil {
				return err
			}

			m := reflect.MakeSlice(reflect.TypeOf(make([]FlowEvent, 0)), len(ee.Value), len(ee.Value))
			for k := range ee.Value {
				e := ee.Value[k]
				eventMix, err := utils.DoubleMarshal[[]ValueItemMix](e.Value)
				if err != nil {
					return err
				}
				fe, err := parseEvent(eventMix)
				if err != nil {
					return err
				}
				m.Index(k).Set(reflect.ValueOf(fe))
			}
			f.Set(m)
		case []Scale:
			arguments, err := utils.DoubleMarshal[[]Argument](v.Value.Value)
			if err != nil {
				return err
			}

			m := reflect.MakeSlice(reflect.TypeOf(make([]Scale, 0)), len(arguments), len(arguments))
			for k := range arguments {
				s := &Scale{}
				argument := arguments[k]
				err = parseValue(argument.Value, s)
				if err != nil {
					return err
				}
				m.Index(k).Set(reflect.ValueOf(*s))
			}
			f.Set(m)
		}

	}

	return nil
}

func parseEvent(eventMix []ValueItemMix) (FlowEvent, error) {
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
			err = parseValue(a, &event)
			if err != nil {
				return event, err
			}
		}
	}
	return event, nil
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
		select (.Name != null and .Name != "getSecret" and .Name != "setupFunction" and .Name != "getFile" ) |
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

// func (c *FlowInformation) GetID() string {
// 	// return c.id
// 	return ""
// }

// func (c *FlowInformation) GetValueHash() string {

// 	// to calculate if the flow needs updating
// 	// the scale has to be removed
// 	b, err := doubleMarshal[interface{}](c.ast)
// 	if err != nil {
// 		slog.Error("error getting value hash", slog.Any("error", err))
// 		return ""
// 	}

// 	// remove all scale attributes from body and definition
// 	r, err := jq[any](`del(.DeclarationList.[].List.[].Initializer.Value[] | select(.Key.Value == "scale")) |
// 	 del(.Body.[].List.[].Initializer.Value[] | select(.Key.Value == "scale"))`, b)
// 	if err != nil {
// 		slog.Error("error parsing value hash", slog.Any("error", err))
// 		return ""
// 	}

// 	if len(r) != 1 {
// 		slog.Error("unexpected results creating hash for flow, expected one result")
// 		return ""
// 	}

// 	h := sha256.New()
// 	s := fmt.Sprintf("%v", r[0])
// 	sh := h.Sum([]byte(s))

// 	return hex.EncodeToString(sh[:32])
// }
