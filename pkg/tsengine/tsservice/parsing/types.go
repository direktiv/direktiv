package parsing

import "encoding/json"

type Root []json.RawMessage

type VarOrFunction struct {
	Var      *VarDeclaration      `json:"Var,omitempty"`      // Var entry
	Function *FunctionDeclaration `json:"Function,omitempty"` // Function entry
}

type VarDeclaration struct {
	Var  int       `json:"Var"`
	List []VarItem `json:"List"`
}

type VarItem struct {
	Target      Target      `json:"Target"`
	Initializer Initializer `json:"Initializer"`
}

type Target struct {
	Name string `json:"Name"`
	Idx  int    `json:"Idx"`
}

type Initializer struct {
	Callee           Callee         `json:"Callee"`
	LeftParenthesis  int            `json:"LeftParenthesis"`
	ArgumentList     []ArgumentList `json:"ArgumentList"`
	RightParenthesis int            `json:"RightParenthesis"`
}

type Callee struct {
	Name string `json:"Name"`
	Idx  int    `json:"Idx"`
}

type ArgumentList struct {
	LeftBrace  int     `json:"LeftBrace"`
	RightBrace int     `json:"RightBrace"`
	Value      []Value `json:"Value"`
}

type Value struct {
	Key      Key     `json:"Key"`
	Kind     string  `json:"Kind"`
	Value    Literal `json:"Value"`
	Computed bool    `json:"Computed"`
}

type Key struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   string `json:"Value"`
}

type Literal struct {
	Idx     int    `json:"Idx"`
	Literal string `json:"Literal"`
	Value   string `json:"Value"`
}

// Struct for function declarations
type FunctionDeclaration struct {
	Function        int           `json:"Function"`
	Name            Target        `json:"Name"`
	ParameterList   ParameterList `json:"ParameterList"`
	Body            Body          `json:"Body"`
	Source          string        `json:"Source"`
	DeclarationList interface{}   `json:"DeclarationList"`
	Async           bool          `json:"Async"`
	Generator       bool          `json:"Generator"`
}

type ParameterList struct {
	Opening int         `json:"Opening"`
	List    interface{} `json:"List"`
	Rest    interface{} `json:"Rest"`
	Closing int         `json:"Closing"`
}

type Body struct {
	LeftBrace  int         `json:"LeftBrace"`
	List       interface{} `json:"List"`
	RightBrace int         `json:"RightBrace"`
}
