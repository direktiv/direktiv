package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/server"
	"github.com/mattn/go-shellwords"
)

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (e Env) toKV() string {
	return fmt.Sprintf("%s=%s", e.Name, e.Value)
}

// nolint
type Command struct {
	Command         string `json:"command"`
	Envs            []Env  `json:"envs"`
	StopOnError     bool   `json:"stop"`
	SuppressCommand bool   `json:"suppress_command"`
	SuppressOutput  bool   `json:"suppress_output"`
}

type Commands struct {
	Commands []Command `json:"commands"`
}

// nolint
type CommandsResponse struct {
	Error  string
	Output interface{}
}

// nolint
func RunCommands(ctx context.Context, in Commands, info *server.ExecutionInfo) (interface{}, error) {
	commandOutput := make([]CommandsResponse, 0)

	info.Log.Log("running %d commands", len(in.Commands))

	for a := range in.Commands {
		command := in.Commands[a]

		// print command
		if !command.SuppressCommand {
			info.Log.Log("running command '%s'", command.Command)
		} else {
			info.Log.Log("running command %d", a)
		}

		// reset binary writer
		info.Log.LogData.Reset()

		// set up logs
		if command.SuppressOutput {
			info.Log.SetWriterState(false)
		}

		err := runCmd(command, info)

		cr := CommandsResponse{
			Output: info.Log.LogData.String(),
		}

		// enable writer again
		info.Log.SetWriterState(true)

		if err != nil {
			info.Log.Log("%s", err.Error())
			cr.Error = err.Error()

			// check if it has to stop here
			if command.StopOnError {
				commandOutput = append(commandOutput, cr)

				break
			}
		}

		commandOutput = append(commandOutput, cr)
	}

	return commandOutput, nil
}

func runCmd(command Command, ei *server.ExecutionInfo) error {
	p := shellwords.NewParser()
	p.ParseEnv = true
	p.ParseBacktick = true

	args, err := p.Parse(command.Command)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("no binary provided")
	}

	// always a binary
	bin := args[0]
	argsIn := []string{}
	if len(args) > 1 {
		argsIn = args[1:]
	}

	cmd := exec.CommandContext(context.Background(), bin, argsIn...)
	cmd.Dir = ei.TmpDir
	cmd.Stdout = ei.Log
	cmd.Stderr = ei.Log

	envs := make([]string, 0)
	envs = append(envs, fmt.Sprintf("HOME=%s", ei.TmpDir))
	envs = append(envs, os.Environ()...)

	for i := range command.Envs {
		envs = append(envs, command.Envs[i].toKV())
	}

	cmd.Env = envs

	return cmd.Run()
}
