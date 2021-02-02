package procutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/tkw1536/procutil/term"
)

// StreamingProcess represents a process that uses a streamer to connect to a remote process.
//
// It automatically initializes a local (tty, pty) pair when needed.
type StreamingProcess struct {
	// environment
	Streamer Streamer
	ctx      context.Context

	// external streams
	stdout, stderr io.ReadCloser
	stdin          io.WriteCloser

	// internal streams
	stdoutTerm, stderrTerm, stdinTerm, ptyTerm *term.Terminal

	// for result handling
	outputErrChan chan error
	inputDoneChan chan struct{}
	restoreTerms  sync.Once

	// for cleanup
	exited bool
}

// StreamingProcess implements the Process interface
func init() {
	var _ Process = (*StreamingProcess)(nil)
}

// Streamer represents a connection to a remote stream.
type Streamer interface {
	fmt.Stringer

	Init(ctx context.Context, Term string, isPty bool) error // init initializes this streamer

	StreamOutput(ctx context.Context, stdout, stderr *os.File, restoreTerms func(), errChan chan error) // stream output streams output to stdout and stderr
	StreamInput(ctx context.Context, stdin *os.File, restoreTerms func(), doneChan chan struct{})       // stream input streams input from stdin

	Attach(ctx context.Context, isPty bool) error             // attach attaches to this stream
	ResizeTo(ctx context.Context, size term.WindowSize) error // resize resizes the remote stream
	Result(ctx context.Context) (int, error)                  // result returns the exit code of the streamed process
	Detach(ctx context.Context) error                         // deteach detaches from this stream
}

// String turns StreamingProcess into a string
func (sp *StreamingProcess) String() string {
	if sp == nil {
		return ""
	}

	return sp.Streamer.String()
}

// Init initializes this StreamingProcess
func (sp *StreamingProcess) Init(ctx context.Context, isTerm bool) error {
	sp.ctx = ctx
	if isTerm {
		return sp.initTerm()
	}

	return sp.initPlain()
}

func (sp *StreamingProcess) initPlain() error {
	var err error

	sp.stdout, sp.stdoutTerm, err = term.NewWritePipe()
	if err != nil {
		return err
	}

	sp.stderr, sp.stderrTerm, err = term.NewWritePipe()
	if err != nil {
		return err
	}

	sp.stdinTerm, sp.stdin, err = term.NewReadPipe()
	if err != nil {
		return err
	}

	return nil
}

func (sp *StreamingProcess) initTerm() error {
	// create a new pty
	pty, tty, err := term.OpenTerminal()
	if err != nil {
		return err
	}

	// store the pty for use in resizing
	sp.ptyTerm = pty

	// standard output is the tty
	sp.stdout = tty.File()
	sp.stdoutTerm = tty

	// standard input is the tty
	sp.stdin = tty.File()
	sp.stdinTerm = tty

	return nil
}

// Stdout returns a pipe to Stdout
func (sp *StreamingProcess) Stdout() (io.ReadCloser, error) {
	return sp.stdout, nil
}

// Stderr returns a pipe to Stderr
func (sp *StreamingProcess) Stderr() (io.ReadCloser, error) {
	return sp.stderr, nil
}

// Stdin returns a pipe to Stdin
func (sp *StreamingProcess) Stdin() (io.WriteCloser, error) {
	return sp.stdin, nil
}

// setRawTerminals sets all the terminals to raw mode
func (sp *StreamingProcess) setRawTerminals() error {
	if err := sp.stdoutTerm.SetRawInput(); err != nil {
		return err
	}

	if err := sp.stderrTerm.SetRawInput(); err != nil {
		return err
	}

	if err := sp.stdoutTerm.SetRawOutput(); err != nil {
		return err
	}

	return nil
}

// restoreTerminals restores all the terminal modes
func (sp *StreamingProcess) restoreTerminals() {
	sp.restoreTerms.Do(func() {
		sp.stdoutTerm.RestoreInput()
		sp.stderrTerm.RestoreInput()
		sp.stdinTerm.RestoreOutput()

		// this check has been adapted from upstream; for some reason they hang on specific platforms
		if in := sp.stdinTerm.File(); in != nil && runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
			in.Close()
		}
	})
}

// Start starts this process
func (sp *StreamingProcess) Start(Term string, resizeChan <-chan term.WindowSize, isPty bool) (*os.File, error) {
	if err := sp.Streamer.Init(sp.ctx, Term, isPty); err != nil {
		return nil, err
	}

	if isPty {
		// start resizing the terminal
		go func() {
			if resizeChan == nil {
				return
			}
			for size := range resizeChan {
				sp.ptyTerm.ResizeTo(size)
				sp.Streamer.ResizeTo(sp.ctx, size)
			}
		}()
	}

	// start streaming
	if err := sp.execAndStream(true); err != nil {
		return nil, err
	}

	// and return
	return sp.ptyTerm.File(), nil
}

func (sp *StreamingProcess) execAndStream(isPty bool) error {

	// set all the streams into raw mode
	if err := sp.setRawTerminals(); err != nil {
		return err
	}

	if err := sp.Streamer.Attach(sp.ctx, isPty); err != nil {
		return err
	}

	// setup channels
	sp.outputErrChan = make(chan error)
	sp.inputDoneChan = make(chan struct{})

	// stream input and ouput
	go sp.Streamer.StreamOutput(sp.ctx, sp.stdoutTerm.File(), sp.stderrTerm.File(), sp.restoreTerminals, sp.outputErrChan)
	go sp.Streamer.StreamInput(sp.ctx, sp.stdinTerm.File(), sp.restoreTerminals, sp.inputDoneChan)

	return nil
}

// waitStreams waits for the streams to finish
func (sp *StreamingProcess) waitStreams() error {
	defer sp.restoreTerminals()

	select {
	case err := <-sp.outputErrChan:
		return err
	case <-sp.inputDoneChan: // wait for output also
		select {
		case err := <-sp.outputErrChan:
			return err
		case <-sp.ctx.Done():
			return sp.ctx.Err()
		}
	case <-sp.ctx.Done():
		return sp.ctx.Err()
	}
}

// Wait waits for the process and returns the exit code
func (sp *StreamingProcess) Wait() (code int, err error) {

	// wait for the streams to close
	if err := sp.waitStreams(); err != nil {
		return 0, err
	}

	// and fetch the result
	code, err = sp.Streamer.Result(sp.ctx)
	if err != nil {
		sp.exited = true
	}
	return
}

// Cleanup cleans up this process, typically to kill it.
func (sp *StreamingProcess) Cleanup() (killed bool) {

	if sp.ptyTerm != nil {
		sp.ptyTerm.Close()
		sp.Streamer.Detach(sp.ctx)
		sp.ptyTerm = nil
	}

	return sp.exited // return if we exited
}
