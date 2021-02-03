// +build !windows

package lowlevel

import mobyterm "github.com/moby/term"

// SetWinsize sets the window size of the terminal referred to by the provided file descriptor.
func SetWinsize(fd FileDescriptor, height, width Size) error {
	return mobyterm.SetWinsize(fd, &mobyterm.Winsize{
		Height: height,
		Width:  width,
	})
}
