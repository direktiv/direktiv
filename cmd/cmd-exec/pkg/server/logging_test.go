package server_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	wgExit := &sync.WaitGroup{}
	wgExit.Add(1)

	var b bytes.Buffer
	s, port := startHttpServer(wgExit, &b)

	loggingText := "Hello World"

	os.Setenv("DIREKTIV_HTTP_BACKEND", fmt.Sprintf("http://localhost:%d", port))
	l := server.NewLogger("123")
	l.Log(loggingText)

	// Log adds \n to log text
	assert.Equal(t, loggingText+"\n", b.String())

	// reste buffer
	b.Reset()

	l.Write([]byte(loggingText))
	assert.Equal(t, loggingText, b.String())

	s.Shutdown(context.Background())
	wgExit.Wait()
}

func startHttpServer(wg *sync.WaitGroup, b *bytes.Buffer) (*http.Server, int) {
	// get a free port, release it and use it
	listener, _ := net.Listen("tcp", ":0")
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	fmt.Println("using port:", port)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	// write the post to the buffer
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "got data\n")
		br, _ := io.ReadAll(r.Body)
		b.Write(br)
		defer r.Body.Close()
	})

	go func() {
		defer wg.Done()
		srv.ListenAndServe()
	}()

	return srv, port
}
