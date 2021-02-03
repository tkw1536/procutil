package procutil

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/tkw1536/procutil/term"
)

// Command is an object that allow interacting with an underlying process.
//
// Unlike a Process, a command is safe to be called repeatedly and by multiple goroutines.
type Command struct {
	Process Process

	m sync.Mutex // m protects all fields below

	state commandState // the current state of the underlying process.

	isPty bool // did the call to init() set up a tty

	waitChan     chan struct{} // closed when waiting is done
	waitExitCode int           // exit code from wait
	waitErr      error         // error from wait

	cleanupOnce sync.Once // used to cleanup once
	cleanupErr  error     // error from cleanup
}

type commandState int

const (
	commandStateDefault commandState = iota
	commandStateInit
	commandStateStart
	commandStateWait
	commandStateDone
)

var errCommandAlreadyInitialized = errors.New("Command: Already initialized")

// Init initializes the underlying process by providing it with an appropriate context.
// Once the context is closed, the command will be killed.
//
// Init must be called once.
func (e *Command) Init(ctx context.Context, isTty bool) error {
	e.m.Lock()
	defer e.m.Unlock()

	if e.state != commandStateDefault { // command already initialized
		return errCommandAlreadyInitialized
	}

	if err := e.Process.Init(ctx, isTty); err != nil {
		return err
	}

	e.state = commandStateInit
	e.isPty = isTty

	return nil
}

var errCommandNotInitialized = errors.New("Command: Not initialized")
var errCommandIsATerminal = errors.New("Command: Is a Terminal")

// Start starts the process and sends output to the provided streams.
//
// Init() must be called before a call to Start().
// If this is not the case, an error is returned.
//
// Start() and StartPty() may only be called once.
// Subsequently calls will produce an error.
func (e *Command) Start(Out, Err io.Writer, In io.Reader) error {
	e.m.Lock()
	defer e.m.Unlock()

	if e.state != commandStateInit {
		return errCommandIsATerminal
	}
	if e.isPty {
		return errCommandIsATerminal
	}

	e.state = commandStateStart

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

var errCommandNotATerminal = errors.New("Command: Not a Terminal")

// StartPty runs this process on the given terminal.
//
// Init() must be called before a call to StartPty().
// If this is not the case, an error is returned.
//
// Start() and StartPty() may only be called once.
// Subsequently calls will produce an error.
func (e *Command) StartPty(tm io.ReadWriter, TERM string, resizeChan <-chan term.WindowSize) error {
	e.m.Lock()
	defer e.m.Unlock()

	if e.state != commandStateInit {
		return errCommandNotInitialized
	}
	if !e.isPty {
		return errCommandNotATerminal
	}

	e.state = commandStateStart

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
	if err := e.wait(); err != nil {
		return 0, err
	}

	// wait for the waiting to finish
	<-e.waitChan
	return e.waitExitCode, e.waitErr
}

var errCommandNotRunning = errors.New("Command: Process is not running")

func (e *Command) wait() error {
	e.m.Lock()
	defer e.m.Unlock()

	if e.state == commandStateWait || e.state == commandStateDone { // already waiting or done
		return nil
	}
	if e.state != commandStateStart {
		return errCommandNotRunning
	}

	e.state = commandStateWait
	e.waitChan = make(chan struct{})
	go e.waiter()

	return nil
}

func (e *Command) waiter() {
	// wait for the process and do some cleanup
	code, err := e.Process.Wait()

	e.m.Lock()
	defer e.m.Unlock()

	e.state = commandStateDone
	e.waitExitCode, e.waitErr = code, err
	go e.Cleanup()
	close(e.waitChan)
}

// Stop stops the underlying process.
// When an underlying process is not running, returns an error.
// When the process has already finished running, returns nil.
func (e *Command) Stop() error {
	e.m.Lock()
	defer e.m.Unlock()

	// ensure that the process is not running
	if e.state != commandStateStart && e.state != commandStateWait {
		return errCommandNotRunning
	}

	// if the process has finished, returns nil.
	if e.state == commandStateDone {
		return nil
	}

	// kill the process
	return e.Process.Stop()
}

// Cleanup cleans up this process.
// Cleanup may be called at any point
func (e *Command) Cleanup() error {
	if err := e.cleanup(); err != nil {
		return err
	}

	e.cleanupOnce.Do(func() {
		e.cleanupErr = e.Process.Cleanup()
	})
	return e.cleanupErr
}

var errCommandRunning = errors.New("Command: Process is running")

func (e *Command) cleanup() error {
	e.m.Lock()
	defer e.m.Unlock()

	if e.state != commandStateDone {
		return errCommandRunning
	}

	return nil
}
