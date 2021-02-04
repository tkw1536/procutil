package procutil

import (
	"io"
	"sync"
	"sync/atomic"
)

// DualCloser is an object that can close it's read and write streams.
type DualCloser interface {
	Close() error      // Close closes the read stream of this object.
	CloseWrite() error // CloseWrite closes the write stream of this object.
}

// NewDualCloser returns a new DualCloser that implements the semantics of closer.
//
// When closer is a DualCloser return it unchanged.
// Otherwise returns a DualCloserWrapper.
func NewDualCloser(closer io.Closer) DualCloser {
	if closer == nil {
		return nil
	}

	if dual, isDualCloser := closer.(DualCloser); isDualCloser {
		return dual
	}

	return &DualCloserWrapper{Closer: closer}
}

// DualCloserWrapper implements DualCloser.
// It wraps an existing io.Closer and calls closer.Close() as soon as both Close() and CloseWrite() have been called.
//
// The Close() and CloseWrite() methods can safely be called concurrently.
//
// A DualCloserWrapper may not be copied.
type DualCloserWrapper struct {
	count             uint32    // count contains the number of functions called.
	close, closeWrite sync.Once // used to ensure that close and closewrite are called once

	// Closer may implement Closer or DualCloser
	Closer io.Closer
}

// Close syncronously calls Closer.Close() as soon as CloseWrite() and Close() have been called.
func (w *DualCloserWrapper) Close() (err error) {
	w.close.Do(func() {
		if atomic.AddUint32(&w.count, 1) == 2 {
			err = w.Closer.Close()
		}
	})
	return
}

// CloseWrite syncronously calls Closer.Close() as soon as CloseWrite() and Close() have been called.
func (w *DualCloserWrapper) CloseWrite() (err error) {
	w.closeWrite.Do(func() {
		if atomic.AddUint32(&w.count, 1) == 2 {
			err = w.Closer.Close()
		}
	})
	return
}
