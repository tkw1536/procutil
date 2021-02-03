package lowlevel

import (
	mobyterm "github.com/moby/term"
)

// TerminalState contains os-specific information about terminal state.
type TerminalState struct {
	state mobyterm.State
}

// SetRawTerminal sets the terminal referred to by fd into raw mode and returns it's previous state for use by ResetTerminal.
func SetRawTerminal(fd FileDescriptor) (state *TerminalState, err error) {
	var s *mobyterm.State
	s, err = mobyterm.SetRawTerminal(fd)
	if s != nil {
		state = &TerminalState{
			state: *s,
		}
	}
	return
}

// SetRawTerminalOutput sets the output of the terminal referred to by fd into raw mode and returns it's previous state for use by ResetTerminal.
func SetRawTerminalOutput(fd FileDescriptor) (state *TerminalState, err error) {
	var s *mobyterm.State
	s, err = mobyterm.SetRawTerminalOutput(fd)
	if s != nil {
		state = &TerminalState{
			state: *s,
		}
	}
	return
}

// ResetTerminal resets the terminal (input or output) mode referred to by fd into the mode described by state.
func ResetTerminal(fd FileDescriptor, state *TerminalState) error {
	var s *mobyterm.State
	if state != nil {
		s = &state.state
	}
	return mobyterm.RestoreTerminal(fd, s)
}
