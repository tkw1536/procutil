package lowlevel

import (
	mobyterm "github.com/moby/term"
)

// FileDescriptor is an os-specific alias for a type representing file descriptors.
type FileDescriptor = uintptr

// GetFdInfo returns information about the terminal referred to by file.
func GetFdInfo(file interface{}) (fd FileDescriptor, isTerminal bool) {
	return mobyterm.GetFdInfo(file)
}
