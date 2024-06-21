package tsservice

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"dario.cat/mergo"
	"github.com/direktiv/direktiv/pkg/tsengine/transpiler"
	"github.com/direktiv/direktiv/pkg/tsengine/tstypes"
	"github.com/dop251/goja"
	"github.com/dop251/goja/ast"
)

type Compiler struct {
	path       string
	namespace  string
	javaScript string
	ast        *ast.Program
}

func NewTSServiceCompiler(namespace, path, script string) (*Compiler, error) {
	tt, err := transpiler.NewTranspiler()
	if err != nil {
		return nil, fmt.Errorf("failed to create transpiler: %w", err)
	}

	// Transpile TypeScript to JavaScript
	js, err := tt.Transpile(script)
	if err != nil {
		return nil, fmt.Errorf("failed to transpile script: %w", err)
	}

	// Parse JavaScript
	parsedAST, err := goja.Parse(path, js)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JavaScript: %w", err)
	}

	// Create map of blacklisted functions
	disallowedFunctions := map[string]bool{
		"eval":        true, // Disallow 'eval' to prevent arbitrary code execution (major security risk).
		"setTimeout":  true, // Disallow 'setTimeout' to avoid unexpected timing issues and potential infinite loops.
		"setInterval": true, // Disallow 'setInterval' for the same reasons as 'setTimeout'.
		"Function":    true, // Disallow the 'Function' constructor to prevent code injection via string-based function creation.
		"require":     true, // Disallow 'require' (if not using a module system) to avoid loading external modules unexpectedly.
	}

	// Validate function calls in global scope, with disallowed functions
	if err := validateGlobalFunctionCalls(parsedAST, disallowedFunctions); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &Compiler{
		path:       path,
		namespace:  namespace,
		javaScript: js,
		ast:        parsedAST,
	}, nil
}

func (c *Compiler) CompileFlow() (*tstypes.FlowInformation, error) {
	flowInfo := &tstypes.FlowInformation{
		Definition: tstypes.DefaultDefinition(),
		Messages:   tstypes.NewMessages(),
		Functions:  make(map[string]tstypes.Function),
		ID:         c.getID(),
	}

	vars := parseTopLevelVarsFromAST(c.ast.Body)
	var err error
	if flowInfo.Definition, err = convertToDefinition(vars); err != nil {
		var noFlowErr *noFlowVariableFoundError
		if errors.As(err, &noFlowErr) {
			slog.Warn("no 'flow' variable found, using default definition")
		} else {
			return nil, fmt.Errorf("error converting to definition: %w", err)
		}
	}

	err = mergo.Merge(flowInfo.Definition, tstypes.DefaultDefinition())
	if err != nil {
		return nil, fmt.Errorf("failed to merge default configuration %w", err)
	}

	// If no state, pick the first function
	if flowInfo.Definition.State == "" {
		if !c.setInitialState(flowInfo) {
			return nil, errors.New("no valid function found to set initial state")
		}
	}

	return flowInfo, nil
}

// parseTopLevelVarsFromAST parses the AST and extracts variable names.
func parseTopLevelVarsFromAST(astBody []ast.Statement) map[string]interface{} {
	slog.Info("starting AST traversal in parseTopLevelVarsFromAST")

	vars := make(map[string]interface{})
	for _, stmt := range astBody {
		slog.Info("processing statement", "type", fmt.Sprintf("%T", stmt))

		if varStmt, ok := stmt.(*ast.VariableStatement); ok {
			for _, binding := range varStmt.List {
				if identifier, ok := binding.Target.(*ast.Identifier); ok {
					initValue := traverseAST(binding.Initializer)
					vars[identifier.Name.String()] = initValue
					slog.Info("Found variable", "name", identifier.Name.String(), "value", initValue)
				}
			}
		}
	}
	slog.Info("completed AST traversal")

	return vars
}

func traverseAST(expr ast.Expression) interface{} {
	slog.Debug("traversing expression", "type", fmt.Sprintf("%T", expr))

	switch v := expr.(type) {
	case *ast.ObjectLiteral:
		objValue := make(map[string]interface{})
		for _, prop := range v.Value {
			if keyedProp, ok := prop.(*ast.PropertyKeyed); ok {
				if key, ok := keyedProp.Key.(*ast.StringLiteral); ok {
					objValue[key.Value.String()] = traverseAST(keyedProp.Value)
					slog.Debug("Found property", "key", key.Value.String())
				}
			}
		}

		return objValue

	case *ast.StringLiteral:
		return v.Value.String()
	case *ast.BooleanLiteral:
		return v.Value
	case *ast.NullLiteral:
		return nil
	case *ast.ArrayLiteral:
		var arrayValue []interface{}
		for _, elem := range v.Value {
			arrayValue = append(arrayValue, traverseAST(elem))
		}

		return arrayValue
	case *ast.NumberLiteral:
		return v.Value
	default:
		slog.Debug("unsupported expression type", "type", fmt.Sprintf("%T", expr))
		return nil
	}
}

func convertToDefinition(vars map[string]interface{}) (*tstypes.Definition, error) {
	for varName, rawValue := range vars {
		if varName == "flow" {
			jsonData, err := json.Marshal(rawValue)
			if err != nil {
				return nil, fmt.Errorf("error marshalling variable 'flow': %w", err)
			}

			var definition tstypes.Definition
			if err := json.Unmarshal(jsonData, &definition); err != nil {
				return nil, fmt.Errorf("error unmarshalling variable 'flow': %w", err)
			}

			return &definition, nil
		}
	}

	return nil, &noFlowVariableFoundError{}
}

type noFlowVariableFoundError struct{}

func (e *noFlowVariableFoundError) Error() string {
	return "no variable named 'flow' found"
}

// validateGlobalFunctionCalls ensures that no function calls other than "setupFunction" are
// made in the global scope. This prevents unexpected side effects like overwriting engine
// variables or functions.
func validateGlobalFunctionCalls(program *ast.Program, disallowedFunctions map[string]bool) error {
	declaredFunctions := make(map[string]bool)

	// First Pass: Gather declared functions
	for _, statement := range program.Body {
		if fnDecl, ok := statement.(*ast.FunctionDeclaration); ok {
			declaredFunctions[fnDecl.Function.Name.Name.String()] = true
			slog.Debug("function declared:", "name", fnDecl.Function.Name.Name.String())
		}
	}

	// Second Pass: Validate function calls
	for _, statement := range program.Body {
		if exprStmt, ok := statement.(*ast.ExpressionStatement); ok {
			if callExpr, ok := exprStmt.Expression.(*ast.CallExpression); ok {
				if identifier, ok := callExpr.Callee.(*ast.Identifier); ok {
					functionName := identifier.Name

					// Skip validation for "setupFunction"
					if functionName == "setupFunction" {
						continue
					}

					// Check if function is declared
					if !declaredFunctions[functionName.String()] {
						slog.Error("undeclared function call in global scope", "function", functionName)
						return fmt.Errorf("function '%s' is not declared", functionName)
					}

					// Check if function is disallowed
					if disallowedFunctions[functionName.String()] {
						slog.Error("disallowed function call in global scope", "function", functionName)
						return fmt.Errorf("function '%s' is not allowed", functionName)
					}
				}
			}
		}
	}

	return nil
}

func (c *Compiler) getID() string {
	str := fmt.Sprintf("%s-%s", c.namespace, c.javaScript)
	hash := sha256.Sum256([]byte(str))

	whitelist := regexp.MustCompile("[^a-zA-Z0-9]+")
	str = whitelist.ReplaceAllString(str, "-")

	// Prevent too long IDs
	if len(str) > 50 {
		str = str[:50]
	}

	return fmt.Sprintf("%s-%x", str, hash[:5])
}

func (c *Compiler) setInitialState(flowInfo *tstypes.FlowInformation) bool {
	for _, statement := range c.ast.Body {
		if fnDecl, ok := statement.(*ast.FunctionDeclaration); ok {
			flowInfo.Definition.State = fnDecl.Function.Name.Name.String()

			return true
		}
	}

	return false
}
