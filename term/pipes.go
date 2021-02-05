package term

import (
	"os"
)

// NewWritePipe returns a new pipe where the write end is a terminal.
func NewWritePipe() (read *os.File, write Terminal, err error) {
	// this function is untested because it is so simple it only relies on os.Pipe.
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return r, NewTerminal(w), nil
}

// NewReadPipe returns a new pipe where the read end is a terminal.
func NewReadPipe() (read Terminal, write *os.File, err error) {
	// this function is untested because it is so simple it only relies on os.Pipe.
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return NewTerminal(r), w, nil
}
