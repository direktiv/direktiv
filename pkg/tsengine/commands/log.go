package commands

import (
	"fmt"
	"log/slog"
)

type logCommand struct {
	attrs []interface{}
}

func NewLogCommand(attrs []interface{}) (logCommand, error) {
	if len(attrs)%2 != 0 {
		return logCommand{}, fmt.Errorf("attrs should be passed as key value pairs.")
	}
	return logCommand{
		attrs: attrs,
	}, nil
}

func (c logCommand) GetName() string {
	return "log"
}

func (c logCommand) GetCommandFunction() interface{} {
	return func(format string, a ...interface{}) {
		slog.Info(fmt.Sprintf(format, a...), c.attrs...)
	}
}
