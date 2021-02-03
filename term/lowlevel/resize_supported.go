// +build !windows

package lowlevel

import (
	"os"
	"os/signal"
	"syscall"
)

// WindowResize returns a channel that receives every time the current terminal window is resized.
// When initial is true, it will additionally receive at some point after the function has returned.
//
// In addition this function also returns a function cleanup that can be used to close the channel notify.
func WindowResize(initial bool) (onResize <-chan struct{}, cleanup func(), err error) {
	c := make(chan struct{}, 1)

	// make a channel for signals and trigger the initial size
	osSigC := make(chan os.Signal, 1)
	signal.Notify(osSigC, syscall.SIGWINCH)

	// every time we get a signal send a trigger
	go func() {
		for range osSigC {
			c <- struct{}{}
		}
	}()

	onResize = c
	cleanup = func() {
		signal.Reset(syscall.SIGWINCH)
		close(c)
	}

	// send initial signal when set
	if initial {
		go func() {
			defer func() { recover() }() // don't care about closing of the channel.
			osSigC <- syscall.SIGWINCH
		}()
	}

	return
}
