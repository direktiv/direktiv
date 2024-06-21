package tsservice

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/direktiv/direktiv/pkg/tsengine/transpiler"
	"github.com/direktiv/direktiv/pkg/tsengine/tstypes"
	"github.com/dop251/goja"
	"github.com/dop251/goja/ast"
)

type Compiler struct {
	// config     Config
	Path       string
	namespace  string
	JavaScript string
	Program    *goja.Program

	ast *ast.Program
}

func NewTSServiceCompiler(namespace, path, script string) (*Compiler, error) {
	tt, err := transpiler.NewTranspiler()
	if err != nil {
		return nil, err
	}

	// make javascript from typescript
	js, err := tt.Transpile(script)
	if err != nil {
		return nil, err
	}

	// check if it is parsable
	ast, err := goja.Parse(path, js)
	if err != nil {
		return nil, err
	}

	// checks if there are function calls in global
	err = validateBodyFunctions(ast)
	if err != nil {
		return nil, err
	}

	// pre compile
	prg, err := goja.Compile(path, js, true)
	if err != nil {
		return nil, err
	}

	return &Compiler{
		Path:       path,
		JavaScript: js,
		ast:        ast,
		Program:    prg,
	}, err
}

func (c *Compiler) getID() string {
	str := fmt.Sprintf("%s-%s", c.namespace, c.JavaScript)
	sh := sha256.Sum256([]byte(str))

	whitelist := regexp.MustCompile("[^a-zA-Z0-9]+")
	str = whitelist.ReplaceAllString(str, "-")

	// Prevent too long ids
	if len(str) > 50 {
		str = str[:50]
	}

	return fmt.Sprintf("%s-%x", str, sh[:5])
}

func (c *Compiler) CompileFlow() (*tstypes.FlowInformation, error) {
	flowInformation := &tstypes.FlowInformation{
		Definition: tstypes.DefaultDefinition(),
		Messages:   tstypes.NewMessages(),
		Functions:  make(map[string]tstypes.Function),
		ID:         c.getID(),
	}
	vars, err := ParseTopLevelVarsFromAST(c.ast.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing AST: %w", err)
	}
	flowInformation.Definition, err = ConvertToDefinition(vars)
	// if no state, we pick the first function
	if flowInformation.Definition.State == "" {
		if !c.selectFirstFunctionState(flowInformation) {
			return nil, fmt.Errorf("no valid function found to set initial state")
		}
	}

	return flowInformation, nil
}

func (c *Compiler) selectFirstFunctionState(flowInformation *tstypes.FlowInformation) bool {
	for _, statement := range c.ast.Body {
		if fnDecl, ok := statement.(*ast.FunctionDeclaration); ok {
			flowInformation.Definition.State = fnDecl.Function.Name.Name.String()
			return true
		}
	}
	return false
}

// ParseTopLevelVarsFromAST parses the AST and extracts variable names.
func ParseTopLevelVarsFromAST(astBody []ast.Statement) (map[string]interface{}, error) {
	slog.Info("Starting AST traversal in ParseTopLevelVarsFromAST") // Start log

	variables := map[string]interface{}{}
	for _, stmt := range astBody {
		slog.Info("Processing statement", "type", fmt.Sprintf("%T", stmt))

		switch s := stmt.(type) {
		case *ast.VariableStatement:
			for _, binding := range s.List {
				switch target := binding.Target.(type) {
				case *ast.Identifier:
					initializerValue := traverseAST(binding.Initializer)
					variables[target.Name.String()] = initializerValue
					slog.Info("Found variable", "name", target.Name.String(), "value", initializerValue)
				}
			}
		}
	}

	slog.Info("Completed AST traversal")
	return variables, nil
}

func traverseAST(expr ast.Expression) interface{} {
	slog.Info("Traversing expression", "type", fmt.Sprintf("%T", expr)) // Log expression type

	switch v := expr.(type) {
	case *ast.ObjectLiteral:
		objectValue := make(map[string]interface{})
		for _, prop := range v.Value {
			switch p := prop.(type) {
			case *ast.PropertyKeyed:
				key, ok := p.Key.(*ast.StringLiteral)
				if !ok {
					continue
				}
				objectValue[key.Value.String()] = traverseAST(p.Value)
				slog.Info("Found property", "key", key.Value.String(), "value", objectValue[key.Value.String()]) // Log property
			}
		}
		return objectValue
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
		slog.Info("Unsupported expression type", "type", fmt.Sprintf("%T", expr)) // Log unsupported types
		return nil
	}
}

func ConvertToDefinition(variables map[string]interface{}) (*tstypes.Definition, error) {
	for varName, rawValue := range variables {
		if varName == "flow" { // Only process variables named "flow"

			// Convert rawValue to JSON bytes
			jsonData, err := json.Marshal(rawValue)
			if err != nil {
				return nil, fmt.Errorf("error marshalling variable 'flow': %w", err)
			}

			slog.Info("Raw JSON data for 'flow':", jsonData) // Log the raw JSON

			// Unmarshal JSON into Definition struct
			var definition tstypes.Definition
			if err := json.Unmarshal(jsonData, &definition); err != nil {
				return nil, fmt.Errorf("error unmarshalling variable 'flow': %w", err)
			}

			slog.Info("Converted to Definition:", definition) // Log the converted struct
			return &definition, nil
		}
	}

	slog.Info("No variable named 'flow' found.") // Log if 'flow' isn't found
	return nil, nil                              // Or return an error if you require 'flow' to be present
}
