package commands

import "fmt"

func Log(format string, a ...any) {
	// TODO: add proper logging layout
	fmt.Println(fmt.Sprintf(format, a...))
}
