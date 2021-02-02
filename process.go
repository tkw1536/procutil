// Package procutil implements wrapper for various processes.
package procutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/tkw1536/procutil/term"
)

// Process is an object that can be executed and has input and output streams.
type Process interface {
	fmt.Stringer

	// Init initializes this process.
	// ctx represents a context that should be used to run the process in.
	// isPty indicates if this process should be run inside a pty and will be passed accordingly to Start().
	Init(ctx context.Context, isPty bool) error

	// Stdout returns the standard output of this process.
	Stdout() (io.ReadCloser, error)
	// Stderr returns the standard error of this process.
	Stderr() (io.ReadCloser, error)
	// Stdin returns the standard input of this process.
	Stdin() (io.WriteCloser, error)

	// Start starts this process and returns a pointer to the pty terminal.
	// Term is the name of tty to run this on. It is typically stored in the 'TERM' env variable.
	// resizeChan is a channel that will resize a WindowSize object everytime the tty is resized.
	// when isPty is true, resizeChan is guaranteed to not be nil.
	// isPty indiccates if the process should be started on a pty. When it is false, Term and resizeChan will be zeroed.
	Start(Term string, resizeChan <-chan term.WindowSize, isPty bool) (*os.File, error)

	// Wait waits for this process to exit and returns the exit code.
	Wait() (int, error)

	// Cleanup is called to clean up this process, typically to kill it and cause Wait() to exit.
	// Cleanup is called exactly once during the liftime of the process; when it is not called between Start() and Wait(), it is called afterwards.
	Cleanup() (killed bool)
}

// Command is a convenience wrapper around a process.
type Command struct {
	Process Process

	isInit bool
	isTty  bool

	cleanupOnce *sync.Once
}

var errCommandNotInitialized = errors.New("Command: Not initialized")
var errCommandAlreadyInitialized = errors.New("Command: Already initialized")

// Init initializes the underlying process.
func (e *Command) Init(ctx context.Context, isTty bool) error {
	if e.isInit {
		return errCommandAlreadyInitialized
	}

	if err := e.Process.Init(ctx, isTty); err != nil {
		return err
	}

	e.isInit = true
	e.isTty = isTty
	e.cleanupOnce = &sync.Once{}

	return nil
}

var errNotATerminal = errors.New("Command: Not a Terminal")
var errIsATerminal = errors.New("Command: Is a Terminal")

// Start starts this process using the provided input and output streams.
func (e *Command) Start(Out, Err io.Writer, In io.Reader) error {
	if !e.isInit {
		return errCommandNotInitialized
	}

	if e.isTty {
		return errIsATerminal
	}

	// fetch all the streams
	stdin, err := e.Process.Stdin()
	if err != nil {
		return err
	}
	stdout, err := e.Process.Stdout()
	if err != nil {
		return err
	}
	stderr, err := e.Process.Stderr()
	if err != nil {
		return err
	}

	// copy over input and output
	go func() {
		defer stdin.Close()
		io.Copy(stdin, In)
	}()

	go func() {
		defer stdout.Close()
		io.Copy(Out, stdout)
	}()

	go func() {
		defer stderr.Close()
		io.Copy(Err, stderr)
	}()

	// Start the process
	_, err = e.Process.Start("", nil, false)
	return err
}

// StartPty runs this process on the given tty
func (e *Command) StartPty(tm io.ReadWriter, TERM string, resizeChan <-chan term.WindowSize) error {
	if !e.isInit {
		return errCommandNotInitialized
	}

	if !e.isTty {
		return errNotATerminal
	}

	// protect against bad callers that pass resizeChan === nil.
	// make a new closed channel.
	if resizeChan == nil {
		c := make(chan term.WindowSize)
		close(c)
		resizeChan = c
	}

	// Start process on the terminal
	f, err := e.Process.Start(TERM, resizeChan, true)
	if err != nil {
		return err
	}

	// start copying both ways and close when done.
	go io.Copy(tm, f)
	go io.Copy(f, tm)

	return nil
}

// Wait waits for this process.
func (e *Command) Wait() (int, error) {
	if !e.isInit {
		return 0, errCommandNotInitialized
	}

	// run process and wait
	defer e.Cleanup()
	return e.Process.Wait()
}

// Cleanup cleans up the underlying process
func (e *Command) Cleanup() error {
	if !e.isInit {
		return errCommandNotInitialized
	}

	// run the cleanup code once
	e.cleanupOnce.Do(func() {
		e.Process.Cleanup()
	})

	return nil
}
