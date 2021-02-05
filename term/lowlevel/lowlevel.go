// Package lowlevel is a wrapper for various lowlevel terminal functionality.
//
// Functions in this package are not intended for usage outside of procutils.
// They are not unit-tested, and may change without notice.
//
// Internally this function is mostly a wrapper around the github.com/creack/pty and github.com/moby/term packages.
// Not all functions are supported on all operating systems.
package lowlevel

import (
	"errors"
	"runtime"
)

// ErrOSUnsupported is returned by various functions to indicate that the operating system is not supported.
var ErrOSUnsupported = errors.New(runtime.GOOS + "not supported")
