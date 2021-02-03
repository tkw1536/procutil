// Package procutil implements wrapper for various processes.
package procutil

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/tkw1536/procutil/term"
)

// Process is an object that can be executed and has input and output streams.
//
// A process may only be used by one caller at the same time; it is not goroutine safe.
// Apart from the String() method, methods may only be called at most once.
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

	// Stop is used to stop a process that is betweeen the start and wait phases.
	Stop() error

	// Wait waits for this process to exit and returns the exit code.
	Wait() (int, error)

	// Cleanup should be called at the end of the lifecyle of the process to clean it up.
	Cleanup() error
}
