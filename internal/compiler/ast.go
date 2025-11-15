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
	Message     string   `json:"message"`
	StartLine   int      `json:"startLine"`
	StartColumn int      `json:"startColumn"`
	EndLine     int      `json:"endLine"`
	EndColumn   int      `json:"endColumn"`
	Severity    Severity `json:"severity"`
}

func (ve *ValidationError) Error() string {
	b, err := json.Marshal(ve)
	if err != nil {
		return fmt.Sprintf("%s (line: %d, column: %d)", ve.Message, ve.StartLine, ve.StartColumn)
	}

	return string(b)
}

type ASTParser struct {
	Script  string
	mapping string

	program *ast.Program
	file    *file.File

	Errors           []*ValidationError
	Actions          []core.ActionConfig
	FlowConfig       core.FlowConfig
	FirstStateFunc   string
	FlowVariable     ast.Expression
	allFunctionNames []string
	allSecretNames   []string
}

func NewASTParser(script, mapping string) (*ASTParser, error) {
	p := &ASTParser{
		file:             file.NewFile("", script, 0),
		Errors:           make([]*ValidationError, 0),
		Actions:          make([]core.ActionConfig, 0),
		Script:           script,
		mapping:          mapping,
		allFunctionNames: make([]string, 0),
		allSecretNames:   make([]string, 0),
	}

	option := parser.WithDisableSourceMaps

	// set mapping if provided
	if mapping != "" {
		sm, err := sourcemap.Parse("", []byte(mapping))
		if err != nil {
			return nil, err
		}
		p.file.SetSourceMap(sm)

		option = parser.WithSourceMapLoader(func(path string) ([]byte, error) {
			return []byte(mapping), nil
		})
	}

	program, err := parser.ParseFile(nil, "", script, 0, option)
	if err != nil {
		return nil, err
	}
	p.program = program

	return p, nil
}

// Parse does everything in a single traversal
func (ap *ASTParser) Parse() error {
	// First pass: collect all function names and find first state function
	for _, stmt := range ap.program.Body {
		if funcDecl, ok := stmt.(*ast.FunctionDeclaration); ok {
			if funcDecl.Function != nil && funcDecl.Function.Name != nil {
				funcName := funcDecl.Function.Name.Name.String()
				ap.allFunctionNames = append(ap.allFunctionNames, funcName)

				// Find first state function
				if ap.FirstStateFunc == "" && strings.HasPrefix(funcName, "state") {
					ap.FirstStateFunc = funcName
				}
			}
		}
	}

	// Second pass: walk through entire tree
	for _, stmt := range ap.program.Body {
		ap.walkNode(stmt, false)
	}

	// Build flow config if flow variable was found
	if ap.FlowVariable != nil {
		config, err := ap.buildFlowConfig()
		if err != nil {
			if vErr, ok := err.(*ValidationError); ok {
				ap.Errors = append(ap.Errors, vErr)
			} else {
				return err
			}
		}
		ap.FlowConfig = config
	} else {
		// No flow variable, use defaults
		ap.FlowConfig = core.FlowConfig{
			Type:    "default",
			Timeout: "PT15M",
			Events:  make([]core.EventConfig, 0),
			State:   ap.FirstStateFunc,
		}
	}

	return nil
}

// walkNode walks through every node in the tree
// isInsideFunc indicates if we're inside a function body
func (ap *ASTParser) walkNode(node ast.Node, isInsideFunc bool) {
	if node == nil {
		return
	}

	// Determine if we're in a state function
	isStateFunc := false
	var funcName string

	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Body {
			ap.walkNode(stmt, false)
		}

	case *ast.FunctionDeclaration:
		if n.Function != nil && n.Function.Name != nil {
			funcName = n.Function.Name.Name.String()
			isStateFunc = strings.HasPrefix(funcName, "state")

			// Validate state function has at least one return
			if isStateFunc {
				hasReturn := ap.checkHasReturn(n.Function.Body)
				if !hasReturn {
					start := ap.file.Position(int(n.Idx0()))
					end := ap.file.Position(int(n.Idx1()))
					ap.Errors = append(ap.Errors, &ValidationError{
						Message:     fmt.Sprintf("state function '%s' must contain at least one return statement with 'transition' or 'finish'", funcName),
						StartLine:   start.Line,
						StartColumn: start.Column,
						EndLine:     end.Line,
						EndColumn:   end.Column,
						Severity:    SeverityError,
					})
				}
			}

			// Walk the function body
			ap.walkFunctionNode(n.Function.Body, isStateFunc)
		}

	case *ast.VariableStatement:
		// Check for flow variable at top level
		if !isInsideFunc {
			for _, b := range n.List {
				if ident, ok := b.Target.(*ast.Identifier); ok && ident.Name == "flow" {
					if ap.FlowVariable == nil {
						ap.FlowVariable = b.Initializer
					}
				}
			}
		}
		// Walk initializers
		for _, decl := range n.List {
			ap.walkNode(decl.Target, isInsideFunc)
			if decl.Initializer != nil {
				ap.walkExpression(decl.Initializer, isInsideFunc, false)
			}
		}

	case *ast.LexicalDeclaration:
		// Check for flow variable at top level
		if !isInsideFunc {
			for _, b := range n.List {
				if ident, ok := b.Target.(*ast.Identifier); ok && ident.Name == "flow" {
					if ap.FlowVariable == nil {
						ap.FlowVariable = b.Initializer
					}
				}
			}
		}
		// Walk initializers
		for _, decl := range n.List {
			ap.walkNode(decl.Target, isInsideFunc)
			if decl.Initializer != nil {
				ap.walkExpression(decl.Initializer, isInsideFunc, false)
			}
		}

	case *ast.ExpressionStatement:
		ap.walkExpression(n.Expression, isInsideFunc, false)

	case *ast.BlockStatement:
		for _, stmt := range n.List {
			ap.walkNode(stmt, isInsideFunc)
		}

	case *ast.IfStatement:
		ap.walkExpression(n.Test, isInsideFunc, false)
		ap.walkNode(n.Consequent, isInsideFunc)
		if n.Alternate != nil {
			ap.walkNode(n.Alternate, isInsideFunc)
		}

	case *ast.ForStatement:
		if n.Initializer != nil {
			ap.walkNode(n.Initializer, isInsideFunc)
		}
		if n.Test != nil {
			ap.walkExpression(n.Test, isInsideFunc, false)
		}
		if n.Update != nil {
			ap.walkExpression(n.Update, isInsideFunc, false)
		}
		ap.walkNode(n.Body, isInsideFunc)

	case *ast.ForInStatement:
		// n.Into is a ForInto interface, not an Expression
		ap.walkExpression(n.Source, isInsideFunc, false)
		ap.walkNode(n.Body, isInsideFunc)

	case *ast.ForOfStatement:
		// n.Into is a ForInto interface, not an Expression
		ap.walkExpression(n.Source, isInsideFunc, false)
		ap.walkNode(n.Body, isInsideFunc)

	case *ast.WhileStatement:
		ap.walkExpression(n.Test, isInsideFunc, false)
		ap.walkNode(n.Body, isInsideFunc)

	case *ast.DoWhileStatement:
		ap.walkNode(n.Body, isInsideFunc)
		ap.walkExpression(n.Test, isInsideFunc, false)

	case *ast.TryStatement:
		ap.walkNode(n.Body, isInsideFunc)
		if n.Catch != nil {
			if n.Catch.Parameter != nil {
				ap.walkNode(n.Catch.Parameter, isInsideFunc)
			}
			ap.walkNode(n.Catch.Body, isInsideFunc)
		}
		if n.Finally != nil {
			ap.walkNode(n.Finally, isInsideFunc)
		}

	case *ast.ThrowStatement:
		ap.walkExpression(n.Argument, isInsideFunc, false)

	case *ast.SwitchStatement:
		ap.walkExpression(n.Discriminant, isInsideFunc, false)
		for _, clause := range n.Body {
			if clause.Test != nil {
				ap.walkExpression(clause.Test, isInsideFunc, false)
			}
			for _, stmt := range clause.Consequent {
				ap.walkNode(stmt, isInsideFunc)
			}
		}

	case *ast.ReturnStatement:
		// This is handled in walkFunctionNode for validation
		if n.Argument != nil {
			ap.walkExpression(n.Argument, isInsideFunc, false)
		}

	case *ast.LabelledStatement:
		ap.walkNode(n.Statement, isInsideFunc)

	case *ast.WithStatement:
		ap.walkExpression(n.Object, isInsideFunc, false)
		ap.walkNode(n.Body, isInsideFunc)
	}
}

// walkFunctionNode walks through function body and validates transition/finish usage
func (ap *ASTParser) walkFunctionNode(node ast.Node, isStateFunc bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.BlockStatement:
		for _, stmt := range n.List {
			ap.walkFunctionNode(stmt, isStateFunc)
		}

	case *ast.IfStatement:
		ap.walkExpression(n.Test, true, isStateFunc)
		ap.walkFunctionNode(n.Consequent, isStateFunc)
		if n.Alternate != nil {
			ap.walkFunctionNode(n.Alternate, isStateFunc)
		}

	case *ast.ReturnStatement:
		// Validate return statements based on function type
		if isStateFunc {
			if !ap.isTransitionCall(n.Argument) {
				start := ap.file.Position(int(n.Idx0()))
				end := ap.file.Position(int(n.Idx1()))
				ap.Errors = append(ap.Errors, &ValidationError{
					Message:     "state function must return a call to 'transition' or 'finish'",
					StartLine:   start.Line,
					StartColumn: start.Column,
					EndLine:     end.Line,
					EndColumn:   end.Column,
					Severity:    SeverityError,
				})
			}
		} else {
			if ap.isTransitionCall(n.Argument) {
				start := ap.file.Position(int(n.Idx0()))
				end := ap.file.Position(int(n.Idx1()))
				ap.Errors = append(ap.Errors, &ValidationError{
					Message:     "non-state function cannot return 'transition' or 'finish'",
					StartLine:   start.Line,
					StartColumn: start.Column,
					EndLine:     end.Line,
					EndColumn:   end.Column,
					Severity:    SeverityError,
				})
			}
		}
		// Don't walk into return argument to avoid double-reporting
		// The validation above already checked what we need

	case *ast.ExpressionStatement:
		ap.walkExpression(n.Expression, true, isStateFunc)

	case *ast.VariableStatement:
		for _, decl := range n.List {
			if decl.Initializer != nil {
				ap.walkExpression(decl.Initializer, true, isStateFunc)
			}
		}

	case *ast.LexicalDeclaration:
		for _, decl := range n.List {
			if decl.Initializer != nil {
				ap.walkExpression(decl.Initializer, true, isStateFunc)
			}
		}

	case *ast.ForStatement:
		if n.Initializer != nil {
			ap.walkFunctionNode(n.Initializer, isStateFunc)
		}
		if n.Test != nil {
			ap.walkExpression(n.Test, true, isStateFunc)
		}
		if n.Update != nil {
			ap.walkExpression(n.Update, true, isStateFunc)
		}
		ap.walkFunctionNode(n.Body, isStateFunc)

	case *ast.ForInStatement:
		// n.Into is a ForInto interface, not an Expression
		ap.walkExpression(n.Source, true, isStateFunc)
		ap.walkFunctionNode(n.Body, isStateFunc)

	case *ast.ForOfStatement:
		// n.Into is a ForInto interface, not an Expression
		ap.walkExpression(n.Source, true, isStateFunc)
		ap.walkFunctionNode(n.Body, isStateFunc)

	case *ast.WhileStatement:
		ap.walkExpression(n.Test, true, isStateFunc)
		ap.walkFunctionNode(n.Body, isStateFunc)

	case *ast.DoWhileStatement:
		ap.walkFunctionNode(n.Body, isStateFunc)
		ap.walkExpression(n.Test, true, isStateFunc)

	case *ast.TryStatement:
		ap.walkFunctionNode(n.Body, isStateFunc)
		if n.Catch != nil {
			ap.walkFunctionNode(n.Catch.Body, isStateFunc)
		}
		if n.Finally != nil {
			ap.walkFunctionNode(n.Finally, isStateFunc)
		}

	case *ast.ThrowStatement:
		if n.Argument != nil {
			ap.walkExpression(n.Argument, true, isStateFunc)
		}

	case *ast.SwitchStatement:
		ap.walkExpression(n.Discriminant, true, isStateFunc)
		for _, clause := range n.Body {
			if clause.Test != nil {
				ap.walkExpression(clause.Test, true, isStateFunc)
			}
			for _, stmt := range clause.Consequent {
				ap.walkFunctionNode(stmt, isStateFunc)
			}
		}

	case *ast.LabelledStatement:
		ap.walkFunctionNode(n.Statement, isStateFunc)

	case *ast.WithStatement:
		ap.walkExpression(n.Object, true, isStateFunc)
		ap.walkFunctionNode(n.Body, isStateFunc)
	}
}

// walkExpression walks through all expressions
// isInsideFunc: true if we're inside any function
// isStateFunc: true if we're inside a state function
func (ap *ASTParser) walkExpression(expr ast.Expression, isInsideFunc bool, isStateFunc bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		// Check for transition/finish calls in non-state functions
		if isInsideFunc && !isStateFunc && ap.isTransitionCall(e) {
			start := ap.file.Position(int(e.Idx0()))
			end := ap.file.Position(int(e.Idx1()))
			ap.Errors = append(ap.Errors, &ValidationError{
				Message:     "non-state function cannot call 'transition' or 'finish'",
				StartLine:   start.Line,
				StartColumn: start.Column,
				EndLine:     end.Line,
				EndColumn:   end.Column,
				Severity:    SeverityError,
			})
		}

		// Get function name from callee (handle both identifier and method calls)
		var funcName string
		var isAllowedTopLevel bool

		if identifier, ok := e.Callee.(*ast.Identifier); ok {
			funcName = identifier.Name.String()
			// Check if this is an allowed top-level function
			isAllowedTopLevel = funcName == "getSecrets" || funcName == "generateAction"

			// Check for generateAction and collect it
			if funcName == "generateAction" {
				if len(e.ArgumentList) == 1 {
					action, err := ap.parseAction(e.ArgumentList[0])
					if err == nil && action.Type == core.FlowActionScopeLocal {
						ap.Actions = append(ap.Actions, action)
					}
				}
			}

			if funcName == "getSecrets" {
				if len(e.ArgumentList) == 1 {
					secrets, err := ap.parseSecrets(e.ArgumentList[0])
					if err != nil {
						start := ap.file.Position(int(e.Idx0()))
						end := ap.file.Position(int(e.Idx1()))
						ap.Errors = append(ap.Errors, &ValidationError{
							Message:     err.Error(),
							StartLine:   start.Line,
							StartColumn: start.Column,
							EndLine:     end.Line,
							EndColumn:   end.Column,
							Severity:    SeverityError,
						})
					}

					ap.allSecretNames = append(ap.allSecretNames, secrets...)
				}
			}
		} else {
			// For method calls, dot expressions, etc., they are NOT allowed at top level
			isAllowedTopLevel = false
		}

		// Check for invalid function calls outside functions
		if !isInsideFunc && !isAllowedTopLevel {
			start := ap.file.Position(int(e.Idx0()))
			end := ap.file.Position(int(e.Idx1()))
			var msg string
			if funcName != "" {
				msg = fmt.Sprintf("function call '%s' is not allowed outside of functions", funcName)
			} else {
				msg = "function call is not allowed outside of functions"
			}
			ap.Errors = append(ap.Errors, &ValidationError{
				Message:     msg,
				StartLine:   start.Line,
				StartColumn: start.Column,
				EndLine:     end.Line,
				EndColumn:   end.Column,
				Severity:    SeverityError,
			})
		}

		// Recurse into arguments to find nested calls
		// Don't recurse into callee to avoid double-reporting
		for _, arg := range e.ArgumentList {
			ap.walkExpression(arg, isInsideFunc, isStateFunc)
		}

	case *ast.ArrayLiteral:
		for _, elem := range e.Value {
			ap.walkExpression(elem, isInsideFunc, isStateFunc)
		}

	case *ast.ObjectLiteral:
		for _, prop := range e.Value {
			ap.walkExpression(prop, isInsideFunc, isStateFunc)
		}

	case *ast.PropertyKeyed:
		ap.walkExpression(e.Key, isInsideFunc, isStateFunc)
		ap.walkExpression(e.Value, isInsideFunc, isStateFunc)

	case *ast.BinaryExpression:
		ap.walkExpression(e.Left, isInsideFunc, isStateFunc)
		ap.walkExpression(e.Right, isInsideFunc, isStateFunc)

	case *ast.UnaryExpression:
		ap.walkExpression(e.Operand, isInsideFunc, isStateFunc)

	case *ast.ConditionalExpression:
		ap.walkExpression(e.Test, isInsideFunc, isStateFunc)
		ap.walkExpression(e.Consequent, isInsideFunc, isStateFunc)
		ap.walkExpression(e.Alternate, isInsideFunc, isStateFunc)

	case *ast.BracketExpression:
		ap.walkExpression(e.Left, isInsideFunc, isStateFunc)
		ap.walkExpression(e.Member, isInsideFunc, isStateFunc)

	case *ast.DotExpression:
		ap.walkExpression(e.Left, isInsideFunc, isStateFunc)

	case *ast.NewExpression:
		ap.walkExpression(e.Callee, isInsideFunc, isStateFunc)
		for _, arg := range e.ArgumentList {
			ap.walkExpression(arg, isInsideFunc, isStateFunc)
		}

	case *ast.SpreadElement:
		ap.walkExpression(e.Expression, isInsideFunc, isStateFunc)

	case *ast.AssignExpression:
		ap.walkExpression(e.Left, isInsideFunc, isStateFunc)
		ap.walkExpression(e.Right, isInsideFunc, isStateFunc)

	case *ast.SequenceExpression:
		for _, expr := range e.Sequence {
			ap.walkExpression(expr, isInsideFunc, isStateFunc)
		}

	case *ast.FunctionLiteral:
		// Anonymous function - treat as regular function (not state function)
		if e.Body != nil {
			ap.walkFunctionNode(e.Body, false)
		}

	case *ast.ArrowFunctionLiteral:
		// Arrow function - treat as regular function (not state function)
		if e.Body != nil {
			// Body could be an expression or a statement
			if blockStmt, ok := e.Body.(ast.Statement); ok {
				ap.walkFunctionNode(blockStmt, false)
			} else if expr, ok := e.Body.(ast.Expression); ok {
				ap.walkExpression(expr, true, false)
			}
		}

	case *ast.TemplateLiteral:
		for _, expr := range e.Expressions {
			ap.walkExpression(expr, isInsideFunc, isStateFunc)
		}
	}
}

// isTransitionCall checks if a node is a call to transition or finish
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

// checkHasReturn recursively checks if a node contains a return statement
func (ap *ASTParser) checkHasReturn(node ast.Node) bool {
	if node == nil {
		return false
	}

	switch n := node.(type) {
	case *ast.ReturnStatement:
		return true
	case *ast.BlockStatement:
		for _, stmt := range n.List {
			if ap.checkHasReturn(stmt) {
				return true
			}
		}
	case *ast.IfStatement:
		if ap.checkHasReturn(n.Consequent) {
			return true
		}
		if n.Alternate != nil && ap.checkHasReturn(n.Alternate) {
			return true
		}
	case *ast.TryStatement:
		if ap.checkHasReturn(n.Body) {
			return true
		}
		if n.Catch != nil && ap.checkHasReturn(n.Catch.Body) {
			return true
		}
		if n.Finally != nil && ap.checkHasReturn(n.Finally) {
			return true
		}
	}

	return false
}

// buildFlowConfig builds the flow configuration from the flow variable
func (ap *ASTParser) buildFlowConfig() (core.FlowConfig, error) {
	flow := core.FlowConfig{
		Type:    "default",
		Timeout: "PT15M",
		Events:  make([]core.EventConfig, 0),
		State:   ap.FirstStateFunc,
	}

	objLit, ok := ap.FlowVariable.(*ast.ObjectLiteral)
	if !ok {
		return flow, nil
	}

	for _, prop := range objLit.Value {
		keyed, ok := prop.(*ast.PropertyKeyed)
		if !ok {
			continue
		}

		var keyName string
		switch k := keyed.Key.(type) {
		case *ast.Identifier:
			keyName = k.Name.String()
		case *ast.StringLiteral:
			keyName = k.Value.String()
		default:
			continue
		}

		switch keyName {
		case "type":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				flowType := strLit.Value.String()
				if !slices.Contains([]string{"default", "cron", "event", "eventsOr", "eventsAnd"}, flowType) {
					start := ap.file.Position(int(keyed.Idx0()))
					end := ap.file.Position(int(keyed.Idx1()))
					return flow, &ValidationError{
						Message:     fmt.Sprintf("unknown flow type: %s", flowType),
						StartLine:   start.Line,
						StartColumn: start.Column,
						EndLine:     end.Line,
						EndColumn:   end.Column,
						Severity:    SeverityError,
					}
				}
				flow.Type = flowType
			}

		case "timeout":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				timeout := strLit.Value.String()
				_, err := duration.Parse(timeout)
				if err != nil {
					start := ap.file.Position(int(keyed.Idx0()))
					end := ap.file.Position(int(keyed.Idx1()))
					return flow, &ValidationError{
						Message:     fmt.Sprintf("invalid timeout pattern '%s', must be ISO8601", timeout),
						StartLine:   start.Line,
						StartColumn: start.Column,
						EndLine:     end.Line,
						EndColumn:   end.Column,
						Severity:    SeverityError,
					}
				}
				flow.Timeout = timeout
			}

		case "cron":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				cronPattern := strLit.Value.String()
				_, err := cron.ParseStandard(cronPattern)
				if err != nil {
					start := ap.file.Position(int(keyed.Idx0()))
					end := ap.file.Position(int(keyed.Idx1()))
					return flow, &ValidationError{
						Message:     fmt.Sprintf("invalid cron pattern: %s", cronPattern),
						StartLine:   start.Line,
						StartColumn: start.Column,
						EndLine:     end.Line,
						EndColumn:   end.Column,
						Severity:    SeverityError,
					}
				}
				flow.Cron = cronPattern
			}

		case "state":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				state := strLit.Value.String()
				if !slices.Contains(ap.allFunctionNames, state) {
					start := ap.file.Position(int(keyed.Idx0()))
					end := ap.file.Position(int(keyed.Idx1()))
					return flow, &ValidationError{
						Message:     fmt.Sprintf("state function '%s' does not exist", state),
						StartLine:   start.Line,
						StartColumn: start.Column,
						EndLine:     end.Line,
						EndColumn:   end.Column,
						Severity:    SeverityError,
					}
				}
				flow.State = state
			}

		case "events":
			if arrLit, ok := keyed.Value.(*ast.ArrayLiteral); ok {
				for _, elem := range arrLit.Value {
					event, err := ap.parseEvent(elem)
					if err == nil {
						flow.Events = append(flow.Events, event)
					}
				}
			}
		}
	}

	// Validate flow configuration consistency
	if flow.Type == "cron" && flow.Cron == "" {
		// Get flow variable location for error reporting
		start := ap.file.Position(0)
		if ap.FlowVariable != nil {
			start = ap.file.Position(int(ap.FlowVariable.Idx0()))
		}
		return flow, &ValidationError{
			Message:     "flow type is 'cron' but no cron pattern is set",
			StartLine:   start.Line,
			StartColumn: start.Column,
			EndLine:     start.Line,
			EndColumn:   start.Column,
			Severity:    SeverityError,
		}
	}

	if (flow.Type == "event" || flow.Type == "eventsOr" || flow.Type == "eventsAnd") && len(flow.Events) == 0 {
		// Get flow variable location for error reporting
		start := ap.file.Position(0)
		if ap.FlowVariable != nil {
			start = ap.file.Position(int(ap.FlowVariable.Idx0()))
		}
		return flow, &ValidationError{
			Message:     "flow type is event-based but no events are defined",
			StartLine:   start.Line,
			StartColumn: start.Column,
			EndLine:     start.Line,
			EndColumn:   start.Column,
			Severity:    SeverityError,
		}
	}

	return flow, nil
}

// parseEvent parses an event configuration
func (ap *ASTParser) parseEvent(expr ast.Expression) (core.EventConfig, error) {
	event := core.EventConfig{
		Context: make(map[string]any),
	}

	objLit, ok := expr.(*ast.ObjectLiteral)
	if !ok {
		return event, fmt.Errorf("event must be an object")
	}

	for _, prop := range objLit.Value {
		keyed, ok := prop.(*ast.PropertyKeyed)
		if !ok {
			continue
		}

		keyLit, ok := keyed.Key.(*ast.StringLiteral)
		if !ok {
			continue
		}

		switch keyLit.Value.String() {
		case "type":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				event.Type = strLit.Value.String()
			}

		case "context":
			if ctxObj, ok := keyed.Value.(*ast.ObjectLiteral); ok {
				for _, ctxProp := range ctxObj.Value {
					if ctxKeyed, ok := ctxProp.(*ast.PropertyKeyed); ok {
						keyStr := ap.extractValue(ctxKeyed.Key)
						valueAny := ap.extractValue(ctxKeyed.Value)
						if keyStr != nil {
							event.Context[fmt.Sprintf("%v", keyStr)] = valueAny
						}
					}
				}
			}
		}
	}

	return event, nil
}

// extractValue extracts a simple value from an expression
func (ap *ASTParser) extractValue(expr ast.Expression) any {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return e.Value.String()
	case *ast.NumberLiteral:
		return e.Value
	case *ast.BooleanLiteral:
		return e.Value
	default:
		// Try JSON marshal/unmarshal as fallback
		type value struct {
			Value any
		}
		var v value
		b, err := json.Marshal(expr)
		if err != nil {
			return nil
		}
		err = json.Unmarshal(b, &v)
		if err != nil {
			return nil
		}
		return v.Value
	}
}

func (ap *ASTParser) parseSecrets(expr ast.Expression) ([]string, error) {
	secrets := make([]string, 0)

	objArray, ok := expr.(*ast.ArrayLiteral)
	if !ok {
		return secrets, fmt.Errorf("secrets must be an array")
	}

	for i := range objArray.Value {
		s := objArray.Value[i]

		sl, ok := s.(*ast.StringLiteral)
		if !ok {
			return secrets, fmt.Errorf("secret values must be a string")
		}

		// fmt.Println(sl.Value)
		// b, _ := json.Marshal(sl)
		// fmt.Println(string(b))
		// fmt.Println(reflect.TypeOf(s))

		secrets = append(secrets, sl.Value.String())
	}

	// arrayList, ok := e.ArgumentList[0].(*ast.ArrayLiteral)
	// if !ok {
	// 	fmt.Println("NOT OKAY")
	// 	start := ap.file.Position(int(e.Idx0()))
	// 	end := ap.file.Position(int(e.Idx1()))
	// 	ap.Errors = append(ap.Errors, &ValidationError{
	// 		Message:     "getSecrets requires a list of secrets",
	// 		StartLine:   start.Line,
	// 		StartColumn: start.Column,
	// 		EndLine:     end.Line,
	// 		EndColumn:   end.Column,
	// 		Severity:    SeverityError,
	// 	})
	// 	// return ?
	// }
	// v, _ := json.Marshal(arrayList)
	// fmt.Println(string(v))

	return secrets, nil
}

// parseAction parses an action configuration from generateAction call
func (ap *ASTParser) parseAction(expr ast.Expression) (core.ActionConfig, error) {
	action := core.ActionConfig{
		Envs: make([]core.EnvironmentVariable, 0),
	}

	objLit, ok := expr.(*ast.ObjectLiteral)
	if !ok {
		return action, fmt.Errorf("action must be an object")
	}

	for _, prop := range objLit.Value {
		keyed, ok := prop.(*ast.PropertyKeyed)
		if !ok {
			continue
		}

		var keyName string
		switch k := keyed.Key.(type) {
		case *ast.Identifier:
			keyName = k.Name.String()
		case *ast.StringLiteral:
			keyName = k.Value.String()
		default:
			continue
		}

		switch keyName {
		case "type":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				action.Type = strLit.Value.String()
			}

		case "image":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				action.Image = strLit.Value.String()
			}

		case "cmd":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				action.Cmd = strLit.Value.String()
			}

		case "size":
			if strLit, ok := keyed.Value.(*ast.StringLiteral); ok {
				action.Size = strLit.Value.String()
			}

		case "envs":
			if arrLit, ok := keyed.Value.(*ast.ArrayLiteral); ok {
				for _, elem := range arrLit.Value {
					if envObj, ok := elem.(*ast.ObjectLiteral); ok {
						var envVar core.EnvironmentVariable
						for _, envProp := range envObj.Value {
							if envKeyed, ok := envProp.(*ast.PropertyKeyed); ok {
								var envKey, envValue string
								if k, ok := envKeyed.Key.(*ast.StringLiteral); ok {
									envKey = k.Value.String()
								}
								if v, ok := envKeyed.Value.(*ast.StringLiteral); ok {
									envValue = v.Value.String()
								}

								if envKey == "name" {
									envVar.Name = envValue
								}
								if envKey == "value" {
									envVar.Value = envValue
								}
							}
						}
						if envVar.Name != "" && envVar.Value != "" {
							action.Envs = append(action.Envs, envVar)
						}
					}
				}
			}
		}
	}

	return action, nil
}
