package lowlevel

import (
	"os"

	mobyterm "github.com/moby/term"
)

// FileDescriptor is an os-specific alias for a type representing file descriptors.
type FileDescriptor = uintptr

// GetFdInfo returns information about the terminal referred to by file.
func GetFdInfo(file *os.File) (fd FileDescriptor, isTerminal bool) {
	return mobyterm.GetFdInfo(file)
}
