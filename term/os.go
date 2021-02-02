package term

import (
	"os"
	"os/exec"
)

// ExecPty starts c on a new pty.
// The user should close pty when finished.
func ExecPty(c *exec.Cmd) (pty *Terminal, err error) {
	fd, err := llStartOnPty(c)
	return NewTerminal(fd), err
}

// GetStdTerminal returns information about the terminal represented by os.Stdout and puts it's input and output in raw mode.
// When os.Stdout is not a terminal, does nothing.
func GetStdTerminal() (term *Terminal, TERM string, resizeChan <-chan WindowSize, cleanup func()) {
	term = NewTerminal(os.Stdout)
	if !term.IsTerminal() { // if we didn't receive a terminal, exit
		term = nil
		return
	}

	term.SetRawInput()
	term.SetRawOutput()

	TERM = os.Getenv("TERM")
	resizeChan = monitorSize(term)

	cleanup = func() {
		term.RestoreInput()
		term.RestoreOutput()
	}

	return
}

func monitorSize(term *Terminal) <-chan WindowSize {
	// send the window size every time we get a resize event
	wsc := make(chan WindowSize, 1)
	go func() {
		for range llWindowResize() {
			size, err := term.GetSize()
			if err != nil || size == nil {
				continue
			}
			wsc <- *size
		}
	}()

	return wsc
}
