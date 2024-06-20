package tsservice

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"

	"dario.cat/mergo"
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
	// check if it is parsable
	ast, err := goja.Parse(path, script)
	if err != nil {
		return nil, err
	}

	// checks if there are function calls in global
	err = validateBodyFunctions(ast)
	if err != nil {
		return nil, err
	}

	// pre compile
	prg, err := goja.Compile(path, script, true)
	if err != nil {
		return nil, err
	}

	return &Compiler{
		Path:       path,
		namespace:  namespace,
		JavaScript: script,
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
		return nil, err // Return nil instead of flowInformation on error
	}

	if err := c.parseAST(astIn, flowInformation); err != nil {
		return nil, err
	}

	// if no state, we pick the first function
	if flowInformation.Definition.State == "" {
		for _, statement := range c.ast.Body {
			if a, ok := statement.(*ast.FunctionDeclaration); ok {
				flowInformation.Definition.State = a.Function.Name.Name.String()
				break
			}
		}
	}

	return flowInformation, nil
}

func (c *Compiler) parseAST(astIn []byte, flowInformation *tstypes.FlowInformation) error {
	dec := json.NewDecoder(bytes.NewReader(astIn))
	for dec.More() {
		var tokenType string
		if err := dec.Decode(&tokenType); err != nil {
			return err
		}

		switch tokenType {
		case "Target":
			var target jsNameStruct
			if err := dec.Decode(&target); err != nil {
				return err
			}
			if target.Name == "flow" {
				if err := handleFlowTarget(dec, flowInformation); err != nil {
					return err
				}
			}
		case "Callee":
			var callee jsNameStruct
			if err := dec.Decode(&callee); err != nil {
				return err
			}
			if callee.Name == "setupFunction" {
				if err := handleSetupFunctionCallee(dec, flowInformation); err != nil {
					return err
				}
			}
			// TODO convert this to middleware with next call
		default:
			// TODO...
		}
	}

	return nil
}

func handleFlowTarget(dec *json.Decoder, flowInformation *tstypes.FlowInformation) error {
	def, err := ParseDefinitionArgs(dec)
	if err != nil {
		return err
	}
	if err := mergo.Merge(flowInformation.Definition, def); err != nil {
		return err
	}
	flowInformation.Messages.Merge(def.Validate())
	flowInformation.Definition = def

	return nil
}

func handleSetupFunctionCallee(dec *json.Decoder, flowInformation *tstypes.FlowInformation) error {
	fn, err := parseCommandArgs(dec)
	if err != nil {
		return err
	}
	flowInformation.Messages.Merge(fn.Validate())
	flowInformation.Functions[fn.GetID()] = *fn

	return nil
}
