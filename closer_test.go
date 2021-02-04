package procutil

import (
	"io"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

type testCloser struct{}

func (testCloser) Close() error { return nil }

type testDualCloser struct{}

func (testDualCloser) Close() error { return nil }

func (testDualCloser) CloseWrite() error { return nil }

func TestNewDualCloser(t *testing.T) {
	type args struct {
		closer io.Closer
	}
	tests := []struct {
		name string
		args args
		want DualCloser
	}{
		{"passing nil returns nil", args{nil}, nil},
		{"passing a closer wraps it", args{closer: testCloser{}}, &DualCloserWrapper{Closer: testCloser{}}},
		{"passing a dualcloser returns it as is", args{closer: testDualCloser{}}, testDualCloser{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDualCloser(tt.args.closer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDualCloser() = %v, want %v", got, tt.want)
			}
		})
	}
}

type closerWithCount struct {
	Effect func()
	count  int64
}

func (c *closerWithCount) Close() error {
	atomic.AddInt64(&c.count, 1)

	if c.Effect != nil {
		c.Effect()
	}

	return nil
}

func (c *closerWithCount) Count() int {
	return int(atomic.LoadInt64(&c.count))
}

func TestDualCloserWrapper(t *testing.T) {
	N := 1000 // number of repeats for various things

	t.Run("sequential test (Close, then CloseWrite)", func(t *testing.T) {

		// variable used to check if the function is indeed blocking.
		var outerCallExited = false

		// make a counter which requires something on c to continue.
		counter := &closerWithCount{Effect: func() {
			if outerCallExited {
				t.Error("Failure: outer call exited")
			}
		}}
		closer := &DualCloserWrapper{Closer: counter}

		// Call Close() twice.
		// Both cases should not call the underlying Close() method and only require some changes.

		if closer.Close() != nil {
			t.Error("closer.Close() did not return nil")
		}
		if closer.Close() != nil {
			t.Error("closer.Close() did not return nil")
		}

		// call CloseWrite() which should call the Close() function.
		// This call must be blockig; otherwise the function would not be able to exit

		if closer.CloseWrite() != nil {
			t.Error("closer.CloseWrite() did not return nil")
		}
		outerCallExited = true

		// check that it was actually called
		want := 1
		got := counter.Count()
		if want != got {
			t.Errorf("Close() called %d time(s), but wanted %d time(s)", got, want)
		}

		// call CloseWrite() and Close() again, neither of which should do anything.
		if closer.CloseWrite() != nil {
			t.Error("closer.CloseWrite() did not return nil")
		}
		if closer.Close() != nil {
			t.Error("closer.Close() did not return nil")
		}
		if closer.Close() != nil {
			t.Error("closer.Close() did not return nil")
		}
	})

	t.Run("sequential test (CloseWrite, then Close)", func(t *testing.T) {

		// variable used to check if the function is indeed blocking.
		var outerCallExited = false

		// make a counter which requires something on c to continue.
		counter := &closerWithCount{Effect: func() {
			if outerCallExited {
				t.Error("Failure: outer call exited")
			}
		}}
		closer := &DualCloserWrapper{Closer: counter}

		// Call CloseWrite() twice.
		// Both cases should not call the underlying Close() method and only require some changes.

		if closer.CloseWrite() != nil {
			t.Error("closer.CloseWrite() did not return nil")
		}
		if closer.CloseWrite() != nil {
			t.Error("closer.CloseWrite() did not return nil")
		}

		// call Close() which should call the Close() function.
		// This call must be blockig; otherwise the function would not be able to exit

		if closer.Close() != nil {
			t.Error("closer.Close() did not return nil")
		}
		outerCallExited = true

		// check that it was actually called
		want := 1
		got := counter.Count()
		if want != got {
			t.Errorf("Close() called %d time(s), but wanted %d time(s)", got, want)
		}

		// call Close() and CloseWrite() again, neither of which should do anything.
		if closer.Close() != nil {
			t.Error("closer.Close() did not return nil")
		}
		if closer.CloseWrite() != nil {
			t.Error("closer.CloseWrite() did not return nil")
		}
		if closer.CloseWrite() != nil {
			t.Error("closer.CloseWrite() did not return nil")
		}
	})

	t.Run("call only Close() repeatedly", func(t *testing.T) {
		counter := &closerWithCount{}
		closer := &DualCloserWrapper{Closer: counter}

		// call close N times in parallel
		wg := &sync.WaitGroup{}
		wg.Add(N)
		for i := 0; i < N; i++ {
			go func() {
				defer wg.Done()
				closer.Close()
			}()
		}
		wg.Wait()

		// check that close was never called
		want := 0
		got := counter.Count()
		if want != got {
			t.Errorf("Close() called %d time(s), but wanted %d time(s)", got, want)
		}
	})

	t.Run("call only CloseWrite() repeatedly", func(t *testing.T) {
		counter := &closerWithCount{}
		closer := &DualCloserWrapper{Closer: counter}

		// call close N times in parallel
		wg := &sync.WaitGroup{}
		wg.Add(N)
		for i := 0; i < N; i++ {
			go func() {
				defer wg.Done()
				closer.CloseWrite()
			}()
		}
		wg.Wait()

		// check that close was never called
		want := 0
		got := counter.Count()
		if want != got {
			t.Errorf("Close() called %d time(s), but wanted %d time(s)", got, want)
		}
	})

	t.Run("call both Close() and CloseWrite() repeatedly", func(t *testing.T) {
		counter := &closerWithCount{}
		closer := &DualCloserWrapper{Closer: counter}

		// call close N times in parallel
		wg := &sync.WaitGroup{}
		wg.Add(2 * N)
		for i := 0; i < N; i++ {
			go func() {
				defer wg.Done()
				closer.CloseWrite()
			}()
			go func() {
				defer wg.Done()
				closer.Close()
			}()
		}
		wg.Wait()

		// check that close was never called
		want := 1
		got := counter.Count()
		if want != got {
			t.Errorf("Close() called %d time(s), but wanted %d time(s)", got, want)
		}
	})
}
