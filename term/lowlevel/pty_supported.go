// +build !windows

package lowlevel

import (
	"os"
	"os/exec"

	creackpty "github.com/creack/pty"
)

// PTYSupport indicates if the current operating system supports OpenPty() and StartOnPty() methods
const PTYSupport = true // platform is supported

// OpenPty opens a new tty and returns the corresponding (tty, pty) file descriptors.
func OpenPty() (tty, pty *os.File, err error) {
	return creackpty.Open()
}

// StartOnPty starts c on a new pty and returns a file descriptor describing it.
func StartOnPty(c *exec.Cmd) (fd *os.File, err error) {
	return creackpty.Start(c)
}
