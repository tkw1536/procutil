package term

import (
	"os"
	"os/exec"

	"github.com/tkw1536/procutil/term/lowlevel"
)

// ExecTerminal starts c on a new pty.
// The user should close pty when finished.
func ExecTerminal(c *exec.Cmd) (pty *Terminal, err error) {
	fd, err := lowlevel.StartOnPty(c)
	return NewTerminal(fd), err
}

// GetStdTerminal returns information about the terminal represented by os.Stdout and puts it's input and output in raw mode.
// When os.Stdout is not a terminal, does nothing.
func GetStdTerminal() (term *Terminal, TERM string, resizeChan <-chan WindowSize, cleanup func(), err error) {
	term = NewTerminal(os.Stdout)
	cleanup = func() {}
	if !term.IsTerminal() { // if we didn't receive a terminal, exit
		term = nil
		return
	}

	err = term.SetRawInput()
	if err != nil {
		return
	}

	err = term.SetRawOutput()
	if err != nil {
		term.RestoreInput() // restore input which we may have broken
		return
	}

	var resizeCleanup func()
	resizeChan, resizeCleanup, err = monitorSize(term)
	if err != nil {
		// restore input and ouput to preven breakage
		term.RestoreInput()
		term.RestoreOutput()
		return
	}

	TERM = os.Getenv("TERM")

	cleanup = func() {
		term.RestoreInput()
		term.RestoreOutput()
		resizeCleanup()
	}

	return
}

func monitorSize(term *Terminal) (ws <-chan WindowSize, cleanup func(), err error) {
	// send the window size every time we get a resize event
	wsc := make(chan WindowSize, 1)
	onResize, cleanup, err := lowlevel.WindowResize(true)
	if err != nil {
		return nil, nil, err
	}

	// whenver we get a signal, get the current terminal size, and return it!
	go func() {
		for range onResize {
			size, err := term.GetSize()
			if err != nil || size == nil {
				continue
			}
			wsc <- *size
		}
	}()

	return wsc, cleanup, nil
}
