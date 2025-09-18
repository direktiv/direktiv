package compiler

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/grafana/sobek/ast"
	"github.com/grafana/sobek/parser"
	"github.com/robfig/cron/v3"
	"github.com/sosodev/duration"
)

type Validator struct {
	errors []string
}

func ValidateTransitions(src, mapping string) ([]string, error) {
	option := parser.WithSourceMapLoader(func(path string) ([]byte, error) {
		return []byte(mapping), nil
	})

	program, err := parser.ParseFile(nil, "", strings.NewReader(src), 0, option)
	if err != nil {
		return nil, err
	}

	validator := &Validator{}
	validator.walk(program, false)

	return validator.errors, nil
}

// walk recursively traverses the AST, applying validation rules based on the current context.
func (v *Validator) walk(node ast.Node, isStateFunc bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Body {
			v.walk(stmt, false)
		}

	case *ast.FunctionDeclaration:
		// Check for the function's name and get its body.
		var funcName string
		// Defensive check: ensure the Function and its Identifier fields are not nil.
		if n.Function != nil && n.Function.Name != nil {
			funcName = n.Function.Name.Name.String()
		}

		isNewStateFunc := strings.HasPrefix(funcName, "state")

		// For state functions, we must ensure at least one return exists.
		if isNewStateFunc {
			hasReturn := v.checkIfHasReturn(n.Function.Body)
			if !hasReturn {
				v.errors = append(v.errors, fmt.Sprintf("state function '%s' must contain at least one return statement", funcName))
			}
		}

		// Continue the main traversal with the new context.
		if n.Function != nil {
			v.walk(n.Function.Body, isNewStateFunc)
		}

	case *ast.BlockStatement:
		if n.List != nil {
			for _, stmt := range n.List {
				v.walk(stmt, isStateFunc)
			}
		}

	case *ast.IfStatement:
		v.walk(n.Consequent, isStateFunc)
		if n.Alternate != nil {
			v.walk(n.Alternate, isStateFunc)
		}

	case *ast.ReturnStatement:
		// Rule 1: A state function must return a transition call.
		if isStateFunc {
			if !v.isTransitionCall(n.Argument) {
				v.errors = append(v.errors, "state function has a return statement that is not a call to 'transition'")
			}
		} else {
			// Rule 2: A non-state function cannot return a transition call.
			if v.isTransitionCall(n.Argument) {
				v.errors = append(v.errors, "non-state function calls 'transition' or 'finish' in its return statement")
			}
		}

	case *ast.CallExpression:
		// Rule 2 (cont.): A non-state function cannot call transition at all.
		if !isStateFunc {
			if v.isTransitionCall(n) {
				v.errors = append(v.errors, "non-state function calls 'transition' or 'finish'.")
			}
		}
		// Continue walking the arguments in case of nested calls

		if n.ArgumentList != nil {
			for _, arg := range n.ArgumentList {
				v.walk(arg, isStateFunc)
			}
		}

	case *ast.ExpressionStatement:
		v.walk(n.Expression, isStateFunc)
	}
}

// isTransitionCall checks if a node represents a function call to "transition".
func (v *Validator) isTransitionCall(node ast.Node) bool {
	callExpr, ok := node.(*ast.CallExpression)
	if !ok || callExpr.Callee == nil {
		return false
	}

	callee, ok := callExpr.Callee.(*ast.Identifier)
	if !ok || callee == nil {
		return false
	}

	return callee.Name == "transition" || callee.Name == "finish"
}

// checkIfHasReturn recursively checks for the existence of a return statement.
func (v *Validator) checkIfHasReturn(node ast.Node) bool {
	if node == nil {
		return false
	}

	switch n := node.(type) {
	case *ast.ReturnStatement:
		return true
	case *ast.BlockStatement:
		for _, stmt := range n.List {
			if v.checkIfHasReturn(stmt) {
				return true
			}
		}
	case *ast.IfStatement:
		if v.checkIfHasReturn(n.Consequent) {
			return true
		}
		if n.Alternate != nil && v.checkIfHasReturn(n.Alternate) {
			return true
		}
	}
	return false
}

func ValidateBody(src, mapping string) error {
	option := parser.WithSourceMapLoader(func(path string) ([]byte, error) {
		return []byte(mapping), nil
	})

	prog, err := parser.ParseFile(nil, "", src, 0, option)
	if err != nil {
		return err
	}

	seenFlow := false
	fns := false

	for _, stmt := range prog.Body {
		switch s := stmt.(type) {
		case *ast.FunctionDeclaration:
			// allowed but set first function
			fns = true
		case *ast.VariableStatement:
			for _, b := range s.List { // here b is *ast.Binding
				ident, ok := b.Target.(*ast.Identifier)
				if !ok {
					return fmt.Errorf("unexpected binding target: %T", b.Target)
				}
				if ident.Name != "flow" {
					return fmt.Errorf("only 'flow' allowed at top-level, got %q", ident.Name)
				}
				if seenFlow {
					return fmt.Errorf("duplicate 'flow' declaration")
				}
				seenFlow = true
			}
		case *ast.EmptyStatement:
			// ignore stray semicolons
		default:
			return fmt.Errorf("top-level %T not allowed", s)
		}
	}

	if !fns {
		return fmt.Errorf("no functions defined")
	}

	return nil
}

func ValidateConfig(src, mapping string) (*core.FlowConfig, error) {
	flow := &core.FlowConfig{
		Type:    "default",
		Timeout: "PT15M",
		Events:  make([]*core.EventConfig, 0),
	}

	option := parser.WithSourceMapLoader(func(path string) ([]byte, error) {
		return []byte(mapping), nil
	})

	prog, err := parser.ParseFile(nil, "", src, 0, option)
	if err != nil {
		return flow, err
	}

	// get all functions for the first function
	functions := []string{}
	for _, stmt := range prog.Body {
		switch s := stmt.(type) {
		case *ast.FunctionDeclaration:
			// set  starting state to the first function
			// will be replace in parsing if set
			if flow.State == "" {
				flow.State = s.Function.Name.Name.String()
			}
			functions = append(functions, s.Function.Name.Name.String())
		}
	}

	for _, stmt := range prog.Body {
		switch s := stmt.(type) {
		case *ast.VariableStatement:
			for _, b := range s.List {
				ident, ok := b.Target.(*ast.Identifier)
				if !ok {
					continue
				}

				literal, ok := b.Initializer.(*ast.ObjectLiteral)
				if !ok {
					continue
				}

				if ident.Name == "flow" {
					for i := range literal.Value {
						p := literal.Value[i]

						literal, ok := p.(*ast.PropertyKeyed)
						if !ok {
							return nil, fmt.Errorf("wrong type in flow config")
						}

						key, ok := literal.Key.(*ast.StringLiteral)
						if !ok {
							return flow, fmt.Errorf("wrong type in flow config")
						}

						switch value := literal.Value.(type) {
						case *ast.StringLiteral:
							err = setAndvalidate(flow, functions, key.Value.String(), value.Value.String())
							if err != nil {
								return flow, err
							}
						case *ast.ArrayLiteral:
							if key.Value.String() == "events" {
								for i := range value.Value {
									e := value.Value[i]
									event, err := processEvent(e)
									if err != nil {

									}
									flow.Events = append(flow.Events, event)
								}
							} else {
								return flow, fmt.Errorf("only events allowed in config")
							}
						default:
							return flow, fmt.Errorf("wrong type in flow config")
						}
					}
				}
			}
		}
	}

	// generic tests
	if flow.Type == cronFlowType && flow.Cron == "" {
		return flow, fmt.Errorf("flow of type cron but no cron set")
	}

	if (flow.Type == eventFlowType || flow.Type == eventsAndFlowType || flow.Type == eventsOrFlowType) &&
		len(flow.Events) == 0 {
		return flow, fmt.Errorf("flow of type event but no events set")
	}

	return flow, nil
}

const (
	defaultFlowType   = "default"
	cronFlowType      = "cron"
	eventFlowType     = "event"
	eventsOrFlowType  = "eventsOr"
	eventsAndFlowType = "eventsAnd"
)

var (
	flowTypes = []string{defaultFlowType, cronFlowType, eventFlowType, eventsOrFlowType, eventsAndFlowType}
)

func processEvent(l ast.Expression) (*core.EventConfig, error) {
	event := &core.EventConfig{
		Context: make(map[string]any),
	}

	e, ok := l.(*ast.ObjectLiteral)
	if !ok {
		return event, fmt.Errorf("wrong format for events")
	}

	for a := range e.Value {
		eventPart := e.Value[a]

		k, ok := eventPart.(*ast.PropertyKeyed)
		if !ok {
			return event, fmt.Errorf("wrong format for events")
		}

		// hashmap so we get the key
		sl, ok := k.Key.(*ast.StringLiteral)
		if !ok {
			return event, fmt.Errorf("wrong format for events")
		}

		switch sl.Value.String() {
		case "type":
			v, ok := k.Value.(*ast.StringLiteral)
			if !ok {
				return event, fmt.Errorf("wrong format for events")
			}

			event.Type = v.Value.String()
		case "context":
			v, ok := k.Value.(*ast.ObjectLiteral)
			if !ok {
				return event, fmt.Errorf("wrong format for events")
			}

			for i := range v.Value {
				kv := v.Value[i]
				pk, ok := kv.(*ast.PropertyKeyed)
				if !ok {
					return event, fmt.Errorf("wrong format for events")
				}

				doubleMarshal := func(in ast.Expression) (any, error) {
					type value struct {
						// Literal string
						Value any
					}
					var v value

					g, err := json.Marshal(in)
					if err != nil {
						return "", err
					}

					err = json.Unmarshal(g, &v)
					return v.Value, err
				}

				key, err := doubleMarshal(pk.Key)
				if err != nil {
					return event, err
				}
				value, err := doubleMarshal(pk.Value)
				if err != nil {
					return event, err
				}

				event.Context[fmt.Sprintf("%v", key)] = value

			}
		default:
			return event, fmt.Errorf("wrong attribute for events")
		}
	}

	return event, nil
}

func setAndvalidate(flow *core.FlowConfig, fns []string, key, value string) error {
	switch key {
	case "type":
		if !slices.Contains(flowTypes, value) {
			return fmt.Errorf("unknown flow type")
		}
		flow.Type = value
	case "timeout":
		_, err := duration.Parse(value)
		if err != nil {
			return fmt.Errorf("invalid pattern for timeout '%s', not ISO8601", value)
		}
		flow.Timeout = value
	case "cron":
		_, err := cron.ParseStandard(value)
		if err != nil {
			return fmt.Errorf("cron pattern %s is invalid", value)
		}
		flow.Cron = value
	case "state":
		if value != "" {
			if !slices.Contains(fns, value) {
				return fmt.Errorf("state %s does not exist", value)
			}
		}
	default:
		return fmt.Errorf("unknown flow attribute %s", key)
	}

	return nil
}
