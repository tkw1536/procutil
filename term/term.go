// Package term contains utilities for dealing with terminals.
package term

import (
	"errors"
	"os"
)

// Terminal represents an interface to a file descriptor that is a terminal
type Terminal struct {
	file *os.File // file

	// returned by llGetFdInfo
	fd         uintptr
	isTerminal bool

	// os-specific state
	inState, outState *llTerminalState
}

// WindowSize represents the size of a terminal window.
type WindowSize struct {
	Height, Width uint16
}

// NewTerminal returns a new terminal instance corresponding to file.
// When file is nil, returns nil.
func NewTerminal(file *os.File) *Terminal {
	if file == nil {
		return nil
	}
	var t Terminal
	t.file = file
	t.fd, t.isTerminal = llGetFdInfo(file)
	return &t
}

// OpenTerminal opens a new terminal and it's corresponding pty
func OpenTerminal() (tty, pty *Terminal, err error) {
	tf, pf, err := llOpenPty()
	return NewTerminal(tf), NewTerminal(pf), err
}

// File returns the os.File corresponding to this terminal, if any.
func (t *Terminal) File() *os.File {
	if t == nil {
		return nil
	}
	return t.file
}

// Close closes the underlying file descriptor, if any.
func (t *Terminal) Close() error {
	if t == nil || t.file == nil {
		return nil
	}
	return t.file.Close()
}

// IsTerminal checks if the underlying file represents a terminal.
func (t *Terminal) IsTerminal() bool {
	return t.isTerminal
}

// SetRawInput sets the input mode of this terminal to raw mode.
// When t is not a terminal, returns nil.
func (t *Terminal) SetRawInput() (err error) {
	if t == nil || !t.isTerminal || t.inState != nil {
		return nil
	}
	t.inState, err = llSetRawTerminal(t.fd)
	return
}

// RestoreInput restores the input mode of this terminal to what it was before the call to SetRawInput().
// When t is not a terminal, or no call to SetRawInput() was made, returns nil.
func (t *Terminal) RestoreInput() error {
	if t == nil || t.inState == nil {
		return nil
	}

	defer func() { t.inState = nil }() // wipe state
	return llResetTerminal(t.fd, t.inState)
}

// SetRawOutput sets the output mode of this terminal to raw mode.
// When t is not a terminal, returns nil.
func (t *Terminal) SetRawOutput() (err error) {
	if t == nil || !t.isTerminal || t.outState != nil {
		return nil
	}
	t.outState, err = llSetRawTerminalOutput(t.fd)
	return
}

// RestoreOutput restores the ouput mode of this terminal to what it was before the call to SetRawOutput().
// When t is not a terminal, or no call to SetRawOutput() was made, returns nil.
func (t *Terminal) RestoreOutput() error {
	if t == nil || t.outState == nil {
		return nil
	}

	defer func() { t.outState = nil }() // wipe state
	return llResetTerminal(t.fd, t.outState)
}

var errNotATerminal = errors.New("GetSize: Not a terminal")

// GetSize returns the current size of this terminal.
// When t is not a terminal, returns an error.
func (t *Terminal) GetSize() (*WindowSize, error) {
	if !t.isTerminal {
		return nil, errNotATerminal
	}

	height, width, err := llGetWinsize(t.fd)
	if err != nil {
		return nil, err
	}

	return &WindowSize{
		Height: height,
		Width:  width,
	}, nil
}

// ResizeTo resizes this terminal to the provided size.
// Errors are silently ignored.
//
// When t does not represent a terminal, does nothing.
func (t *Terminal) ResizeTo(size WindowSize) {
	if !t.isTerminal {
		return
	}

	llSetWinsize(t.fd, size.Height, size.Width)
}
