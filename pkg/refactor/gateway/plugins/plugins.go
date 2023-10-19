package plugins

var registry = make(map[string]string)

func GetSchema(key string) (string, bool) {
	t, b := registry[key]

	return t, b
}

func init() { //nolint
	registry["test"] = "hi"
}
