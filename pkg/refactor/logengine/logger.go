package logengine

type logger interface {
	// logs the msg. Msg can contain string format specifiers.
	// Values passed via a will be applied to the format specifiers if present.
	// Tags will be associated with the log-entry in the logs.
	Log(tags map[string]interface{}, level string, msg string, a ...interface{})
}
