package tsserver

import (
	"github.com/direktiv/direktiv/pkg/tsengine"
)

func RunApplication() {
	srv, err := tsengine.NewServer()
	if err != nil {
		panic(err)
	}

	panic(srv.Start())
}
