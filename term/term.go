// Package term contains utilities for dealing with terminals.
package term

import (
	"errors"
	"io"

	"github.com/tkw1536/procutil/term/lowlevel"
)

// Terminal represents an object that has an underlying read-writer that is potentially a tty.
type Terminal interface {
	// ReadWriteCloser returns the io.ReadWriteCloser corresponding to this terminal, if any.
	ReadWriteCloser() io.ReadWriteCloser

	// Close closes the underlying file descriptor, if any.
	Close() error

	// IsTerminal checks if the underlying file represents a terminal.
	IsTerminal() bool

	// SetRawInput sets the input mode of this terminal to raw mode.
	// When t is not a terminal, returns nil.
	SetRawInput() error

	// RestoreInput restores the input mode of this terminal to what it was before the call to SetRawInput().
	// When t is not a terminal, or no call to SetRawInput() was made, returns nil.
	RestoreInput() error

	// SetRawOutput sets the output mode of this terminal to raw mode.
	// When t is not a terminal, returns nil.
	SetRawOutput() error

	// RestoreOutput restores the ouput mode of this terminal to what it was before the call to SetRawOutput().
	// When t is not a terminal, or no call to SetRawOutput() was made, returns nil.
	RestoreOutput() error

	// GetSize returns the current size of this terminal.
	// When t is not a terminal, returns ErrNotATerminal.
	GetSize() (*WindowSize, error)

	// ResizeTo resizes this terminal to the provided size.
	// Errors are silently ignored.
	//
	// When t does not represent a terminal, returns ErrNotATerminal.
	ResizeTo(size WindowSize) error
}

// ErrNotATerminal is returned when the underlying terminal is not a terminal
var ErrNotATerminal = errors.New("Terminal: File() is not a terminal")

// OpenTerminal opens a new terminal and it's corresponding pty.
func OpenTerminal() (tty, pty Terminal, err error) {
	tf, pf, err := lowlevel.OpenPty()
	return NewTerminal(tf), NewTerminal(pf), err
}

// WindowSize represents the size of a terminal window.
type WindowSize struct {
	Height, Width lowlevel.Size
}

// NewTerminal returns a new terminal instance corresponding to file.
// rwcloser may be nil.
func NewTerminal(rwcloser io.ReadWriteCloser) Terminal {
	if rwcloser == nil {
		return nilTerminal{}
	}
	var t fileTerminal
	t.rwcloser = rwcloser
	t.fd, t.isTerminal = lowlevel.GetFdInfo(rwcloser)
	return &t
}

// fileTerminal represents an interface to a file descriptor that is a terminal
// It implements Terminal.
type fileTerminal struct {
	rwcloser io.ReadWriteCloser // file

	// returned by llGetFdInfo
	fd         lowlevel.FileDescriptor
	isTerminal bool

	// os-specific state
	inState, outState *lowlevel.TerminalState
}

func (t *fileTerminal) ReadWriteCloser() io.ReadWriteCloser {
	return t.rwcloser
}

func (t *fileTerminal) Close() error {
	if t.rwcloser == nil {
		return nil
	}
	return t.rwcloser.Close()
}

func (t *fileTerminal) IsTerminal() bool {
	return t.isTerminal
}

func (t *fileTerminal) SetRawInput() (err error) {
	if !t.isTerminal || t.inState != nil {
		return nil
	}
	t.inState, err = lowlevel.SetRawTerminal(t.fd)
	return
}

func (t *fileTerminal) RestoreInput() error {
	if t.inState == nil {
		return nil
	}

	defer func() { t.inState = nil }() // wipe state
	return lowlevel.ResetTerminal(t.fd, t.inState)
}

func (t *fileTerminal) SetRawOutput() (err error) {
	if !t.isTerminal || t.outState != nil {
		return nil
	}
	t.outState, err = lowlevel.SetRawTerminalOutput(t.fd)
	return
}

func (t *fileTerminal) RestoreOutput() error {
	if t.outState == nil {
		return nil
	}

	defer func() { t.outState = nil }() // wipe state
	return lowlevel.ResetTerminal(t.fd, t.outState)
}

func (t *fileTerminal) GetSize() (*WindowSize, error) {
	if !t.IsTerminal() {
		return nil, ErrNotATerminal
	}

	height, width, err := lowlevel.GetWinsize(t.fd)
	if err != nil {
		return nil, err
	}

	return &WindowSize{
		Height: height,
		Width:  width,
	}, nil
}

func (t *fileTerminal) ResizeTo(size WindowSize) error {
	if !t.IsTerminal() {
		return ErrNotATerminal
	}

	return lowlevel.SetWinsize(t.fd, size.Height, size.Width)
}

// nilTerminal implements Terminal returns a negative result for every command
type nilTerminal struct{}

func (nilTerminal) ReadWriteCloser() io.ReadWriteCloser { return nil }
func (nilTerminal) Close() error                        { return nil }
func (nilTerminal) IsTerminal() bool                    { return false }
func (nilTerminal) SetRawInput() error                  { return nil }
func (nilTerminal) RestoreInput() error                 { return nil }
func (nilTerminal) SetRawOutput() error                 { return nil }
func (nilTerminal) RestoreOutput() error                { return nil }
func (nilTerminal) GetSize() (*WindowSize, error)       { return nil, ErrNotATerminal }
func (nilTerminal) ResizeTo(size WindowSize) error      { return ErrNotATerminal }
