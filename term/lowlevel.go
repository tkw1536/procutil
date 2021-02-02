// +build !windows

package term

// This file contains wrappers for low-level, os-specific APIs.
// Dummy implementations can be found in lowlevel_unsupported.go; these are to prevent compilation issues on Windows.

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/moby/term"
)

// llWindowResize sends a signal every time the terminal window resizes
func llWindowResize() <-chan struct{} {
	c := make(chan struct{}, 1)

	// make a channel for signals and trigger the initial size
	osSigC := make(chan os.Signal, 1)
	signal.Notify(osSigC, syscall.SIGWINCH)
	defer func() { osSigC <- syscall.SIGWINCH }()

	// every time we get a signal send a trigger
	go func() {
		for range osSigC {
			c <- struct{}{}
		}
	}()

	return c
}

func llGetWinsize(fd uintptr) (height, width uint16, err error) {
	size, err := term.GetWinsize(fd)
	if err != nil {
		return 0, 0, err
	}
	return size.Height, size.Width, nil
}

func llSetWinsize(fd uintptr, height, width uint16) error {
	return term.SetWinsize(fd, &term.Winsize{
		Height: height,
		Width:  width,
	})
}

var llGetFdInfo = term.GetFdInfo

type llTerminalState = term.State

var llSetRawTerminal = (func(uintptr) (*llTerminalState, error))(term.SetRawTerminal)
var llSetRawTerminalOutput = (func(uintptr) (*llTerminalState, error))(term.SetRawTerminalOutput)
var llResetTerminal = (func(uintptr, *llTerminalState) error)(term.RestoreTerminal)

var llStartOnPty = pty.Start
var llOpenPty = pty.Open
