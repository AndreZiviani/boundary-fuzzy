package run

import (
	"context"
	"io"
	"os/exec"
)

type Task struct {
	Command string
	Args    []string
	Output  io.Reader
	Cmd     *exec.Cmd

	Cancel func()
	DoneCh chan struct{}
	CmdErr error
}

func RunTask(command string, args []string) *Task {
	ctx, cancel := context.WithCancel(context.Background())

	// The command you want to run along with the argument
	cmd := exec.CommandContext(ctx, command, args...)

	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan struct{})

	t := Task{
		Command: command,
		Args:    args,
		Output:  r,
		Cmd:     cmd,
		Cancel:  cancel,
		DoneCh:  done,
	}

	// Goroutine that runs the Task
	go func() {
		cmd.Run()
		<-done
	}()

	// Goroutine that waits for the command to exit using cmd.Wait().
	// It closes the doneCh to indicate to users of Daemon that
	// the command has finished.
	//
	// This goroutine must be run only after cmd.Start() returns,
	// otherwise cmd.Run() may panic.
	//
	// The command can exit either:
	// * normally with success;
	// * with failure due to a command error; or
	// * with failure due to Context cancelation.
	go func() {
		err := cmd.Wait()
		t.CmdErr = err
		close(t.DoneCh)
	}()

	return &t

}
