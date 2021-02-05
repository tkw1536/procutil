package procutil

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/tkw1536/procutil/term"
)

// testProcess is a process used for testing.
// It plainly records all calls to it.
type testProcess struct {
	Out string          // output for out
	Err string          // output for stderr
	in  closeableBuffer // buffer to read stdin from

	// channels that are closed when any of the streams are closed.
	outChan chan struct{}
	errChan chan struct{}
	inChan  chan struct{}

	ExitCode int // ExitCode to return

	pty, tty term.Terminal

	InitCalled    bool
	StopCalled    bool
	CleanupCalled bool
}

type testProcessClosable struct {
	c chan struct{}
}

func (c testProcessClosable) Close() error {
	if c.c != nil {
		close(c.c)
	}
	return nil
}

type closeableBuffer struct {
	*bytes.Buffer
	testProcessClosable
}

type closeableReader struct {
	*strings.Reader
	testProcessClosable
}

func (tp *testProcess) String() string {
	return "TestProcess"
}

func (tp *testProcess) Init(ctx context.Context, isPty bool) error {
	tp.InitCalled = true

	tp.outChan = make(chan struct{})
	tp.errChan = make(chan struct{})
	tp.inChan = make(chan struct{})

	return nil
}

func (tp *testProcess) Stdout() (io.ReadCloser, error) {
	return &closeableReader{
		Reader:              strings.NewReader(tp.Out),
		testProcessClosable: testProcessClosable{tp.outChan},
	}, nil
}

func (tp *testProcess) Stderr() (io.ReadCloser, error) {

	return &closeableReader{
		Reader:              strings.NewReader(tp.Err),
		testProcessClosable: testProcessClosable{tp.errChan},
	}, nil
}

func (tp *testProcess) Stdin() (io.WriteCloser, error) {

	tp.in = closeableBuffer{
		Buffer:              new(bytes.Buffer),
		testProcessClosable: testProcessClosable{tp.inChan},
	}
	return &tp.in, nil
}

func (tp *testProcess) Start(Term string, resizeChan <-chan term.WindowSize, isPty bool) (term.Terminal, error) {
	if isPty == true {
		panic("not supported")
	}

	return nil, nil
}

func (tp *testProcess) Stop() error {
	tp.StopCalled = true
	return nil
}

func (tp *testProcess) Wait() (int, error) {
	<-tp.inChan
	<-tp.outChan
	<-tp.errChan

	return tp.ExitCode, nil
}

func (tp *testProcess) Cleanup() error {
	tp.CleanupCalled = true
	if tp.pty != nil {
		tp.pty.Close()
	}
	return nil
}

func TestCommandNoTty(t *testing.T) {

	tests := []struct {
		name                     string
		wantOut, wantErr, wantIn string
		wantStop                 bool
		wantCode                 int
	}{
		{
			name:     "output without error",
			wantOut:  "output",
			wantErr:  "error",
			wantIn:   "input",
			wantCode: 0,
		},

		{
			name:     "output with error",
			wantOut:  "output",
			wantErr:  "error",
			wantIn:   "input",
			wantCode: 1,
		},

		{
			name:     "output without error but with stop",
			wantOut:  "output",
			wantErr:  "error",
			wantIn:   "input",
			wantStop: true,
			wantCode: 0,
		},

		{
			name:     "output with error and with stop",
			wantOut:  "output",
			wantErr:  "error",
			wantIn:   "input",
			wantStop: true,
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup test process
			process := &testProcess{
				Out:      tt.wantOut,
				Err:      tt.wantErr,
				ExitCode: tt.wantCode,
			}

			//
			// go through the normal process lifecyle
			//

			command := &Command{
				Process: process,
			}

			if got := command.Init(nil, false); got != nil {
				t.Error("Command.Init() didn't return nil")
			}

			var outBuffer, errBuffer bytes.Buffer
			inBuffer := strings.NewReader(tt.wantIn)

			if got := command.Start(&outBuffer, &errBuffer, inBuffer); got != nil {
				t.Error("Command.Start() didn't return nil")
			}

			if tt.wantStop {
				if got := command.Stop(); got != nil {
					t.Error("Command.Stop() didn't return nil")
				}
			}

			if gotCode, gotErr := command.Wait(); gotCode != tt.wantCode || gotErr != nil {
				t.Errorf("Command.Wait() didn't return (%d, nil)", tt.wantCode)
			}

			if got := command.Cleanup(); got != nil {
				t.Error("Command.Cleanup() didn't return nil")
			}

			//
			// ensure that we got the right output and input
			//

			if outBuffer.String() != tt.wantOut {
				t.Error("Command didn't write output")
			}

			if errBuffer.String() != tt.wantErr {
				t.Error("Command didn't write error")
			}

			if process.in.String() != tt.wantIn {
				t.Error("Command didn't read input")
			}

			//
			// ensure that all the internal methods were called
			//

			if !process.InitCalled {
				t.Error("Command didn't call Init()")
			}

			if process.StopCalled != tt.wantStop {
				t.Error("Command StopCalled != wantStop")
			}

			if !process.CleanupCalled {
				t.Error("Command didn't call Cleanup()")
			}

		})
	}
}
