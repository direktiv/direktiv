package commands

import "strings"

type TrimCommand struct{}

func (c *TrimCommand) GetName() string {
	return "trim"
}

func (c *TrimCommand) GetCommandFunction() interface{} {
	return func(in string) string {
		return strings.TrimSuffix(in, "\n")
	}
}
