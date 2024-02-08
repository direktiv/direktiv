package commands_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/commands"
	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/server"
	"github.com/stretchr/testify/assert"
)

var b bytes.Buffer

func TestMain(m *testing.M) {
	s, port, wgExit := startHttpServer(&b)
	os.Setenv("DIREKTIV_HTTP_BACKEND", fmt.Sprintf("http://localhost:%d", port))

	m.Run()

	s.Shutdown(context.Background())
	wgExit.Wait()
}

func TestCommandSuppress(t *testing.T) {
	b.Reset()

	dir, _ := os.MkdirTemp(os.TempDir(), "prefix")
	defer os.RemoveAll(dir)

	se := &server.ExecutionInfo{
		TmpDir: dir,
		Log:    server.NewLogger("123"),
	}

	cmds := commands.Commands{
		[]commands.Command{
			{
				Command: "ls",
			},
		},
	}

	commands.RunCommands(context.Background(), cmds, se)

	// prints command
	assert.True(t, strings.HasPrefix(b.String(), "running 1 commands\nrunning command 'ls'"))

	b.Reset()

	cmds = commands.Commands{
		[]commands.Command{
			{
				Command:         "ls",
				SuppressCommand: true,
			},
		},
	}

	commands.RunCommands(context.Background(), cmds, se)

	// suppress command
	assert.True(t, strings.HasPrefix(b.String(), "running 1 commands\nrunning command 0"))
}

func TestOutputSuppress(t *testing.T) {
	b.Reset()

	dir, _ := os.MkdirTemp(os.TempDir(), "prefix")
	defer os.RemoveAll(dir)

	se := &server.ExecutionInfo{
		TmpDir: dir,
		Log:    server.NewLogger("123"),
	}

	cmds := commands.Commands{
		[]commands.Command{
			{
				Command:        "echo jens",
				SuppressOutput: true,
			},
		},
	}

	commands.RunCommands(context.Background(), cmds, se)
	assert.Equal(t, "running 1 commands\nrunning command 'echo jens'\n", b.String())

	// run unsuppressed
	cmds.Commands[0].SuppressOutput = false
	b.Reset()
	commands.RunCommands(context.Background(), cmds, se)

	assert.Equal(t, "running 1 commands\nrunning command 'echo jens'\njens\n", b.String())
}

func TestErrors(t *testing.T) {
	b.Reset()

	dir, _ := os.MkdirTemp(os.TempDir(), "prefix")
	defer os.RemoveAll(dir)

	se := &server.ExecutionInfo{
		TmpDir: dir,
		Log:    server.NewLogger("123"),
	}

	cmds := commands.Commands{
		[]commands.Command{
			{
				Command:        "onetwo",
				SuppressOutput: true,
			},
		},
	}

	commands.RunCommands(context.Background(), cmds, se)
	assert.Equal(t, "running 1 commands\nrunning command 'onetwo'\nexec: \"onetwo\": executable file not found in $PATH\n", b.String())

	// run unsuppressed
	cmds.Commands[0].SuppressOutput = false
	b.Reset()
	commands.RunCommands(context.Background(), cmds, se)

	// prints error message, although suppressed
	assert.Equal(t, "running 1 commands\nrunning command 'onetwo'\nexec: \"onetwo\": executable file not found in $PATH\n", b.String())
}

func TestStopOnTestErrors(t *testing.T) {
	b.Reset()

	dir, _ := os.MkdirTemp(os.TempDir(), "prefix")
	se := &server.ExecutionInfo{
		TmpDir: dir,
		Log:    server.NewLogger("123"),
	}
	defer os.RemoveAll(dir)

	cmds := commands.Commands{
		[]commands.Command{
			{
				Command:     "echo hello1",
				StopOnError: true,
			},
			{
				Command:     "does not exist",
				StopOnError: true,
			},
			{
				Command:     "echo hello2",
				StopOnError: true,
			},
		},
	}

	commands.RunCommands(context.Background(), cmds, se)
	assert.Equal(t, "running 3 commands\nrunning command 'echo hello1'\nhello1\nrunning command 'does not exist'\nexec: \"does\": executable file not found in $PATH\n", b.String())

	cmds.Commands[1].StopOnError = false

	b.Reset()
	commands.RunCommands(context.Background(), cmds, se)
	assert.Equal(t, "running 3 commands\nrunning command 'echo hello1'\nhello1\nrunning command 'does not exist'\nexec: \"does\": executable file not found in $PATH\nrunning command 'echo hello2'\nhello2\n", b.String())
}

func startHttpServer(b *bytes.Buffer) (*http.Server, int, *sync.WaitGroup) {
	wgExit := &sync.WaitGroup{}
	wgExit.Add(1)

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
		defer wgExit.Done()
		srv.ListenAndServe()
	}()

	return srv, port, wgExit
}

func TestCommandEnvs(t *testing.T) {
	b.Reset()

	dir, _ := os.MkdirTemp(os.TempDir(), "prefix")
	defer os.RemoveAll(dir)

	se := &server.ExecutionInfo{
		TmpDir: dir,
		Log:    server.NewLogger("123"),
	}

	cmds := commands.Commands{
		[]commands.Command{
			{
				Command: "bash -c 'env | grep HELLO'",
				Envs: []commands.Env{
					{
						Name:  "HELLO",
						Value: "WORLD",
					},
				},
			},
		},
	}

	commands.RunCommands(context.Background(), cmds, se)
	assert.Contains(t, b.String(), "HELLO=WORLD")
}
