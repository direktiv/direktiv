package commands

import "fmt"

type LogCommand struct{}

func (c *LogCommand) GetName() string {
	return "log"
}

func (c *LogCommand) GetCommandFunction() interface{} {
	return func(format string, a ...interface{}) {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}
