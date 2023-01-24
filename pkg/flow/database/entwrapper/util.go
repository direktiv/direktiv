package entwrapper

import (
	"runtime"
	"strings"
)

func parent() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}
