package procutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkw1536/procutil/term"
)

// ExecProcess ia a process that is executed via a syscall to exec().
// It implements the Process interface.
type ExecProcess struct {
	Command string   // command to run
	Args    []string // arguments for the command
	Workdir string   // workding directory of the process, defaults to ""
	Env     []string // Environment variables of the form "KEY=VALUE"

	cmd *exec.Cmd // command being run
}

// ExecProcess implements the Process interface
func init() {
	var _ Process = (*ExecProcess)(nil)
}

// Init initializes this process
func (sp *ExecProcess) Init(ctx context.Context, isPty bool) error {
	// exec.Command internally does use LookPath(), but doesn't return an error
	// Instead we explicitly call LookPath() to intercept the error

	exe, err := exec.LookPath(sp.Command)
	if err != nil {
		err = errors.Wrapf(err, "Can't find %s in path", sp.Command)
		return err
	}

	sp.cmd = exec.Command(exe, sp.Args...)
	sp.cmd.Dir = sp.Workdir
	sp.cmd.Env = sp.Env

	return nil
}

// String turns ShellProcess into a string
func (sp *ExecProcess) String() string {
	if sp == nil || sp.cmd == nil {
		return ""
	}

	return strings.Join(append([]string{sp.cmd.Path}, sp.cmd.Args...), " ")
}

// Stdout returns a pipe to Stdout
func (sp *ExecProcess) Stdout() (io.ReadCloser, error) {
	return sp.cmd.StdoutPipe()
}

// Stderr returns a pipe to Stderr
func (sp *ExecProcess) Stderr() (io.ReadCloser, error) {
	return sp.cmd.StderrPipe()
}

// Stdin returns a pipe to Stdin
func (sp *ExecProcess) Stdin() (io.WriteCloser, error) {
	return sp.cmd.StdinPipe()
}

// Start starts this process
func (sp *ExecProcess) Start(Term string, resizeChan <-chan term.WindowSize, isPty bool) (*os.File, error) {
	// not a pty => start the process and be done!
	if !isPty {
		return nil, sp.cmd.Start()
	}

	// add the terminal environment variable
	sp.cmd.Env = append(sp.cmd.Env, fmt.Sprintf("TERM=%s", Term))

	// start the pty
	t, err := term.ExecTerminal(sp.cmd)
	if err != nil {
		return nil, err
	}

	// start tracking window size
	go func() {
		for size := range resizeChan {
			t.ResizeTo(size)
		}
	}()

	// and return a function for this
	return t.File(), nil
}

// Wait waits for the process and returns the exit code
func (sp *ExecProcess) Wait() (code int, err error) {
	// wait for the command
	err = sp.cmd.Wait()
	code = 255

	// if we have a failure and it's not an exit code
	// we need to return an error
	_, isExitError := err.(*exec.ExitError)
	if err != nil && !isExitError {
		err = errors.Wrap(err, "cmd.Wait() returned non-exit-error")
		return
	}

	// return the exit code
	code = sp.cmd.ProcessState.ExitCode()
	return code, nil
}

var errExecStopFailure = errors.New("ExecProcess: Failed to kill process")

// Stop is used to stop a running process.
func (sp *ExecProcess) Stop() (err error) {
	// silence any panic()ing errors, but return false!
	defer func() {
		if err == nil {
			recover()
			err = errExecStopFailure
		}
	}()

	// kill the process, and prevent further attempts
	return sp.cmd.Process.Kill()
}

// Cleanup cleans up this process, typically killing it
func (sp *ExecProcess) Cleanup() error {
	sp.cmd.Process = nil // remove the process object
	return nil
}
