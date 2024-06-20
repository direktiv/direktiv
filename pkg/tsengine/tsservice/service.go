package tsservice

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/direktiv/direktiv/pkg/tsengine/transpiler"
	"github.com/direktiv/direktiv/pkg/tsengine/tsservice/parsing"
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

	astIn, err := json.MarshalIndent(c.ast.Body, "", "   ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling AST: %w", err)
	}

	if err := c.parseAST(astIn, flowInformation); err != nil {
		return nil, fmt.Errorf("error parsing AST: %w", err)
	}

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

func (c *Compiler) parseAST(astIn []byte, flowInformation *tstypes.FlowInformation) error {
	var root parsing.Root
	if err := json.Unmarshal(astIn, &root); err != nil {
		return fmt.Errorf("unmarshaling AST: %w", err)
	}

	// Iterate over each RawMessage in the root slice
	for _, rawEntry := range root {
		var entry parsing.VarOrFunction
		if err := json.Unmarshal(rawEntry, &entry); err != nil {
			return fmt.Errorf("unmarshaling entry: %w", err)
		}

		if entry.Var != nil {
			// Handle the VarDeclaration (entry.Var)
		} else if entry.Function != nil {
			// Handle the FunctionDeclaration (entry.Function)
			funcName := entry.Function.Name.Name
			entry.Function.ParameterList.List
		}
	}

	return nil
}
