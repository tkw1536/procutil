// +build windows

package lowlevel

import (
	"os"
	"os/exec"
)

// OpenPty opens a new tty and returns the corresponding (tty, pty) file descriptors.
func OpenPty() (tty, pty *os.File, err error) {
	return nil, nil, ErrWindowsUnsupported
}

// StartOnPty starts c on a new pty and returns a file descriptor describing it.
func StartOnPty(c *exec.Cmd) (fd *os.File, err error) {
	return nil, ErrWindowsUnsupported
}
