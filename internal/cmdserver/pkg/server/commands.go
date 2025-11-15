package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/direktiv/direktiv/internal/telemetry"
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

// nolint
type CommandsResponse struct {
	Error  string
	Output interface{}
	Stdout string `json:"-"`
}

// nolint
func RunCommands(ctx context.Context, in Payload, info *ExecutionInfo) (interface{}, error) {
	commandOutput := make([]CommandsResponse, 0)

	slog.Info("starting to run commands", "total", len(in.Commands))

	telemetry.LogInstance(ctx, telemetry.LogLevelInfo, fmt.Sprintf("running %d commands", len(in.Commands)))

	for a := range in.Commands {
		command := in.Commands[a]

		if !command.SuppressCommand {
			info.Log.Logf("running command '%s'", command.Command)
			slog.Info("running command", "index", a, "command", command.Command)
		} else {
			info.Log.Logf("running command %d", a)
			slog.Info("running command with suppressed output", "index", a)
		}

		// reset binary writer
		info.Log.LogData.Reset()
		slog.Debug("resetting log data buffer", "index", a)

		// suppress output if needed
		if command.SuppressOutput {
			info.Log.SetWriterState(false)
			slog.Debug("suppressing output for command", "index", a)
		}

		err := runCmd(command, info)

		cr := CommandsResponse{
			Stdout: info.Log.LogData.String(),
		}

		// enable writer again
		info.Log.SetWriterState(true)
		slog.Debug("re-enabling log writer after command", "index", a)

		if err != nil {
			info.Log.Logf("%s", err.Error())
			slog.Error("command execution failed", "index", a, "error", err)

			cr.Error = err.Error()
			if command.StopOnError {
				info.Log.Logf("stopping execution due to error in command %d", a)
				slog.Warn("execution stopped because of failure", "index", a)
				commandOutput = append(commandOutput, cr)
				return commandOutput, fmt.Errorf("stopped because command %d failed", a)
			}
		}

		// copy stdout from command in output
		cr.Output = cr.Stdout

		// if valid json make it a "real" json in the response
		if json.Valid([]byte(cr.Stdout)) {
			var raw json.RawMessage
			json.Unmarshal([]byte(cr.Stdout), &raw)
			cr.Output = raw
		}

		commandOutput = append(commandOutput, cr)
	}

	slog.Info("finished running commands", "total", len(in.Commands))
	return commandOutput, nil
}

func runCmd(command Command, ei *ExecutionInfo) error {
	slog.Debug("parsing command", "command", command.Command)

	p := shellwords.NewParser()
	p.ParseEnv = true
	p.ParseBacktick = true

	args, err := p.Parse(command.Command)
	if err != nil {
		slog.Error("failed to parse command", "command", command.Command, "error", err)
		return err
	}

	if len(args) == 0 {
		slog.Error("no binary provided in command", "command", command.Command)
		return fmt.Errorf("no binary provided")
	}

	// extract binary and arguments
	bin := args[0]
	argsIn := []string{}
	if len(args) > 1 {
		argsIn = args[1:]
	}

	slog.Info("executing command", "binary", bin, "args", argsIn)

	cmd := exec.CommandContext(context.Background(), bin, argsIn...)
	cmd.Dir = ei.TmpDir
	cmd.Stdout = ei.Log
	cmd.Stderr = ei.Log

	// set environment variables
	envs := make([]string, 0)
	envs = append(envs, fmt.Sprintf("HOME=%s", ei.TmpDir))
	envs = append(envs, os.Environ()...)

	for i := range command.Envs {
		env := command.Envs[i].toKV()
		envs = append(envs, env)
		slog.Debug("adding environment variable", "env", env)
	}

	cmd.Env = envs

	slog.Debug("starting command execution", "binary", bin)
	err = cmd.Run()
	if err != nil {
		slog.Error("command execution failed", "binary", bin, "error", err)
		return err
	}

	slog.Info("command executed successfully", "binary", bin)

	return nil
}
