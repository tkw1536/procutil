// +build windows

package lowlevel

// TODO: At the moment most of these functions return an error on windows.

// WindowResize returns a channel that receives every time the current terminal window is resized.
// When initial is true, it will additionally receive at some point after the function has returned.
//
// In addition this function also returns a function cleanup that can be used to close the channel notify.
func WindowResize(initial bool) (onResize <-chan struct{}, cleanup func(), err error) {
	c := make(chan struct{})

	// on windows, only send an initial signal and do not listen to resize events (for now!)
	if initial {
		go func() {
			defer func() { recover() }() // don't care about closing the channel.
			c <- struct{}{}
		}()
	}

	return c, func() { close(c) }, nil
}
