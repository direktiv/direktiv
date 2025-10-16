package compiler

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/go-sourcemap/sourcemap"
	"github.com/grafana/sobek/ast"
	"github.com/grafana/sobek/file"
	"github.com/grafana/sobek/parser"
	"github.com/robfig/cron/v3"
	"github.com/sosodev/duration"
)

type Severity string

const (
	SeverityHint    Severity = "hint"
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type ValidationError struct {
	Message        string `json:"message"`
	StartLine      int    `json:"startLine"`
	StartColumn    int    `json:"startColumn"`
	EndLine        int    `json:"endLine"`
	EndColumn      int    `json:"endColumn"`
	Severity    	 Severity `json:"severity"`
}

func (ve *ValidationError) Error() string {
	b, err := json.Marshal(ve)
	if err != nil {
		return fmt.Sprintf("%s (line: %d, column: %d)", ve.Message, ve.StartLine, ve.StartColumn)
	}
	return string(b)
}

type ASTParser struct {
	Script, mapping string

	program *ast.Program
	file    *file.File

	Errors  []*ValidationError
	Actions []core.ActionConfig
}

type Validator struct {
	errors []string
}

func NewASTParser(script, mapping string) (*ASTParser, error) {
	p := &ASTParser{
		file:    file.NewFile("", script, 0),
		Errors:  make([]*ValidationError, 0),
		Actions: make([]core.ActionConfig, 0),
		Script:  script,
		mapping: mapping,
	}

	if mapping != "" {
		sm, err := sourcemap.Parse("", []byte(mapping))
		if err != nil {
			return nil, err
		}
		p.file.SetSourceMap(sm)
	}

	option := parser.WithSourceMapLoader(func(path string) ([]byte, error) {
		return []byte(mapping), nil
	})

	program, err := parser.ParseFile(nil, "", script, 0, option)
	if err != nil {
		return nil, err
	}
	p.program = program

	return p, nil
}

func (ap *ASTParser) ValidateTransitions() {
	ap.walk(ap.program, false)
}

// walk recursively traverses the AST, applying validation rules based on the current context.
func (ap *ASTParser) walk(node ast.Node, isStateFunc bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Body {
			ap.walk(stmt, false)
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
			hasReturn := ap.checkIfHasReturn(n.Function.Body)
			if !hasReturn {
				start := ap.file.Position(int(n.Idx0()))
				end := ap.file.Position(int(n.Idx1()))

				ap.Errors = append(ap.Errors, &ValidationError{
					Message: fmt.Sprintf("state function '%s' must contain at least one return statement 'transition' or 'finish'", funcName),
					StartLine:    start.Line,
					StartColumn:  start.Column,
					EndLine: end.Line,
					EndColumn: end.Column,
					Severity: SeverityError,
				})
			}
		}

		// Continue the main traversal with the new context.
		if n.Function != nil {
			ap.walk(n.Function.Body, isNewStateFunc)
		}

	case *ast.BlockStatement:
		if n.List != nil {
			for _, stmt := range n.List {
				ap.walk(stmt, isStateFunc)
			}
		}

	case *ast.IfStatement:
		ap.walk(n.Consequent, isStateFunc)
		if n.Alternate != nil {
			ap.walk(n.Alternate, isStateFunc)
		}

	case *ast.ReturnStatement:
		// Rule 1: A state function must return a transition call.
		if isStateFunc {
			if !ap.isTransitionCall(n.Argument) {
				start := ap.file.Position(int(n.Idx0()))
				end := ap.file.Position(int(n.Idx1()))

				ap.Errors = append(ap.Errors, &ValidationError{
					Message: "state function has a return statement that is not a call to 'transition' or 'finish'",
					StartLine:    start.Line,
					StartColumn:  start.Column,
					EndLine: end.Line,
					EndColumn: end.Column,
					Severity: SeverityError,
				})
			}
		} else {
			// Rule 2: A non-state function cannot return a transition call.
			if ap.isTransitionCall(n.Argument) {
				start := ap.file.Position(int(n.Idx0()))
				end := ap.file.Position(int(n.Idx1()))

				ap.Errors = append(ap.Errors, &ValidationError{
					Message: "non-state function calls 'transition' or 'finish' in its return statement",
					StartLine:    start.Line,
					StartColumn:  start.Column,
					EndLine: end.Line,
					EndColumn: end.Column,
					Severity: SeverityError,
				})
			}
		}

	case *ast.CallExpression:
		// Rule 2 (cont.): A non-state function cannot call transition at all.
		if !isStateFunc {
			if ap.isTransitionCall(n) {
				start := ap.file.Position(int(n.Idx0()))
				end := ap.file.Position(int(n.Idx1()))
				ap.Errors = append(ap.Errors, &ValidationError{
					Message: "non-state function calls 'transition' or 'finish'.",
					StartLine:    start.Line,
					StartColumn:  start.Column,
					EndLine: end.Line,
					EndColumn: end.Column,
					Severity: SeverityError,
				})
			}
		}

		// Continue walking the arguments in case of nested calls
		if n.ArgumentList != nil {
			for _, arg := range n.ArgumentList {
				ap.walk(arg, isStateFunc)
			}
		}

	case *ast.ExpressionStatement:
		ap.walk(n.Expression, isStateFunc)
	}
}

// isTransitionCall checks if a node represents a function call to "transition".
func (ap *ASTParser) isTransitionCall(node ast.Node) bool {
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
func (ap *ASTParser) checkIfHasReturn(node ast.Node) bool {
	if node == nil {
		return false
	}

	switch n := node.(type) {
	case *ast.ReturnStatement:
		return true
	case *ast.BlockStatement:
		for _, stmt := range n.List {
			if ap.checkIfHasReturn(stmt) {
				return true
			}
		}
	case *ast.IfStatement:
		if ap.checkIfHasReturn(n.Consequent) {
			return true
		}
		if n.Alternate != nil && ap.checkIfHasReturn(n.Alternate) {
			return true
		}
	}

	return false
}

func (ap *ASTParser) ValidateConfig() (*core.FlowConfig, error) {
	flow := &core.FlowConfig{
		Type:    "default",
		Timeout: "PT15M",
		Events:  make([]*core.EventConfig, 0),
	}

	// get all functions for the first function
	functions := []string{}
	for _, stmt := range ap.program.Body {
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

	for _, stmt := range ap.program.Body {
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

				// nolint:nestif
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
							err := setAndvalidate(flow, functions, key.Value.String(), value.Value.String())


							if err != nil {
								start := ap.file.Position(int(literal.Key.Idx0()))
								end := ap.file.Position(int(literal.Value.Idx1()))
								ap.Errors = append(ap.Errors, &ValidationError{
									Message:     err.Error(),
									StartLine:   start.Line,
									StartColumn: start.Column,
									EndLine:     end.Line,
									EndColumn:   end.Column,
									Severity:    SeverityError,
								})
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

var flowTypes = []string{defaultFlowType, cronFlowType, eventFlowType, eventsOrFlowType, eventsAndFlowType}

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

func (ap *ASTParser) ValidateFunctionCalls() {
	for i := range ap.program.Body {
		ap.inspectStatement(ap.program.Body[i])
	}
}

func (ap *ASTParser) inspectStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VariableStatement:
		for _, decl := range s.List {
			if decl.Initializer != nil {
				ap.inspectExpression(decl.Initializer)
			}
		}
	case *ast.ExpressionStatement:
		ap.inspectExpression(s.Expression)
	case *ast.FunctionDeclaration:
		// allowed
	case *ast.LexicalDeclaration:
		for _, decl := range s.List {
			if decl.Initializer != nil {
				ap.inspectExpression(decl.Initializer)
			}
		}
	default:
		// we let it through
	}
}

func (ap *ASTParser) parseAction(expr ast.Expression) (core.ActionConfig, error) {
	action := core.ActionConfig{
		Envs: make(map[string]string),
	}

	// Cast to ObjectLiteral
	objLit, ok := expr.(*ast.ObjectLiteral)
	if !ok {
		return core.ActionConfig{}, fmt.Errorf("expected ObjectLiteral, got %T", expr)
	}

	// Iterate through properties
	for _, prop := range objLit.Value {
		// Get the key name from the property
		var keyName string
		if ident, ok := prop.(*ast.PropertyKeyed); ok {
			switch k := ident.Key.(type) {
			case *ast.Identifier:
				keyName = k.Name.String()
			case *ast.StringLiteral:
				keyName = k.Value.String()
			default:
				continue
			}

			switch keyName {
			case "type":
				if strLit, ok := ident.Value.(*ast.StringLiteral); ok {
					action.Type = strLit.Value.String()
				}

			case "size":
				if strLit, ok := ident.Value.(*ast.StringLiteral); ok {
					action.Size = strLit.Value.String()
				}

			case "image":
				if strLit, ok := ident.Value.(*ast.StringLiteral); ok {
					action.Image = strLit.Value.String()
				}

			case "envs":
				if objLit, ok := ident.Value.(*ast.ObjectLiteral); ok {
					for a := range objLit.Value {
						if ident, ok := objLit.Value[a].(*ast.PropertyKeyed); ok {
							var mapKey, mapValue string
							if key, ok := ident.Key.(*ast.StringLiteral); ok {
								mapKey = key.Value.String()
							}

							if value, ok := ident.Value.(*ast.StringLiteral); ok {
								mapValue = value.Value.String()
							}

							if mapKey != "" && mapValue != "" {
								action.Envs[mapKey] = mapValue
							} else {
								start := ap.file.Position(int(expr.Idx0()))
								end := ap.file.Position(int(expr.Idx1()))
								ap.Errors = append(ap.Errors, &ValidationError{
									Message: "generateAction environment varariables have non-string keys or values",
									StartLine:    start.Line,
									StartColumn:  start.Column,
									EndLine: end.Line,
									EndColumn: end.Column,
									Severity: SeverityError,
								})
							}
						}
					}
				}
				if arrLit, ok := ident.Value.(*ast.ArrayLiteral); ok {
					// Parse array elements as key-value pairs
					for i := 0; i < len(arrLit.Value); i += 2 {
						if i+1 < len(arrLit.Value) {
							if keyIdent, ok := arrLit.Value[i].(*ast.Identifier); ok {
								if valIdent, ok := arrLit.Value[i+1].(*ast.Identifier); ok {
									action.Envs[keyIdent.Name.String()] = valIdent.Name.String()
								}
							}
						}
					}
				}
			}
		}
	}

	return action, nil
}

func (ap *ASTParser) inspectExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		msg := "function calls not allowed outside of functions"
		identifier, ok := e.Callee.(*ast.Identifier)
		if ok {
			msg = fmt.Sprintf("function call '%s' is not allowed outside of functions", identifier.Name)
			start := ap.file.Position(int(e.Idx0()))
			end := ap.file.Position(int(e.Idx0()))

			if identifier.Name == "generateAction" {
				// still check for functions
				for a := range e.ArgumentList {
					ap.inspectExpression(e.ArgumentList[a])
				}

				if len(e.ArgumentList) != 1 {
					ap.Errors = append(ap.Errors, &ValidationError{
						Message: "generateAction has no or more than one configuration",
						StartLine:    start.Line,
						StartColumn:  start.Column,
						EndLine: end.Line,
						EndColumn: end.Column,
						Severity: SeverityError,
					})

					return
				}

				action, err := ap.parseAction(e.ArgumentList[0])
				if err != nil {
					ap.Errors = append(ap.Errors, &ValidationError{
						Message: "generateAction has no or more than one configuration",
						StartLine:    start.Line,
						StartColumn:  start.Column,
						EndLine: end.Line,
						EndColumn: end.Column,
						Severity: SeverityError,
					})

					return
				}

				// add them to an action list
				if action.Type == core.FlowActionScopeLocal {
					ap.Actions = append(ap.Actions, action)
				}

				return
			}

			// 'secrets' is allowed but we need to check the params
			if identifier.Name == "getSecrets" {
				for a := range e.ArgumentList {
					ap.inspectExpression(e.ArgumentList[a])
				}

				return
			}
		}

				start := ap.file.Position(int(e.Idx0()))
				end := ap.file.Position(int(e.Idx1()))
		ap.Errors = append(ap.Errors, &ValidationError{
			Message: msg,
			StartLine:    start.Line,
			StartColumn:  start.Column,
			EndLine: end.Line,
			EndColumn: end.Column,
			Severity: SeverityError,
		})
	case *ast.Identifier:
		// allowed
	case *ast.StringLiteral:
		// allowed
	case *ast.NumberLiteral:
		// allowed
	case *ast.SpreadElement:
		ap.inspectExpression(e.Expression)
	case *ast.PropertyKeyed:
		ap.inspectExpression(e.Key)
		ap.inspectExpression(e.Value)
	case *ast.ArrayLiteral:
		for _, elem := range e.Value {
			ap.inspectExpression(elem)
		}
	case *ast.BracketExpression:
		ap.inspectExpression(e.Left)
		ap.inspectExpression(e.Member)
	case *ast.ObjectLiteral:
		for _, prop := range e.Value {
			ap.inspectExpression(prop)
		}
	case *ast.BinaryExpression:
		ap.inspectExpression(e.Left)
		ap.inspectExpression(e.Right)
	case *ast.UnaryExpression:
		ap.inspectExpression(e.Operand)
	case *ast.ConditionalExpression:
		ap.inspectExpression(e.Alternate)
		ap.inspectExpression(e.Consequent)
		ap.inspectExpression(e.Test)
	case *ast.DotExpression:
		ap.inspectExpression(e.Left)
	case *ast.NewExpression:
		for i := range e.ArgumentList {
			ap.inspectExpression(e.ArgumentList[i])
		}
	default:
		// we allowed it by default
	}
}
