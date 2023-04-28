package internallogger

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Error Level = "error"
	Panic Level = "panic"
)
