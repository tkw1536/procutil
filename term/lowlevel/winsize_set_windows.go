// +build windows

package lowlevel

// SetWinsize sets the window size of the terminal referred to by the provided file descriptor.
func SetWinsize(fd FileDescriptor, height, width Size) error {
	return ErrWindowsUnsupported
}
