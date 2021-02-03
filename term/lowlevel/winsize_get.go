package lowlevel

import mobyterm "github.com/moby/term"

// Size is an os-specific alias for dimensions of a terminal.
// It is guaranteed to be some integer type.
type Size = uint16

// GetWinsize gets the window size of the terminal referred to by the provided file descriptor.
func GetWinsize(fd FileDescriptor) (height, width Size, err error) {
	size, err := mobyterm.GetWinsize(fd)
	if err != nil {
		return 0, 0, err
	}
	return size.Height, size.Width, nil
}
