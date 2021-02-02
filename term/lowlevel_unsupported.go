// +build windows

package term

import (
	"errors"
	"os"
	"os/exec"
)

// osWindowResize sends a signal every time the terminal window resizes
func llWindowResize() <-chan struct{} {
	return nil
}

var errOSUnsupported = errors.New("term: Unsupported OS")

func llGetWinsize(fd uintptr) (height, width uint16, err error) {
	return 0, 0, errOSUnsupported
}

func llSetWinsize(fd uintptr, height, width uint16) error {
	return errOSUnsupported
}

func llGetFdInfo(file *os.File) (uintptr, bool) {
	return uintptr(0), false
}

type llTerminalState = struct{}

func llSetRawTerminal(fd uintptr) (*llTerminalState, error) {
	return nil, errOSUnsupported
}

func llSetRawTerminalOutput(fd uintptr) (*llTerminalState, error) {
	return nil, errOSUnsupported
}

func llResetTerminal(fd uintptr, state *llTerminalState) error {
	return errOSUnsupported
}

func llStartOnPty(c *exec.Cmd) (*os.File, error) {
	return nil, errOSUnsupported
}

func llOpenPty() (tty, pty *os.File, error) {
	return nil, nil, errOSUnsupported
}