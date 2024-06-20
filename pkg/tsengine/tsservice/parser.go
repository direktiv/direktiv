package tsservice

import (
	"fmt"

	"github.com/dop251/goja/ast"
)

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
