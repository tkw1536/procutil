// Package term contains utilities for dealing with terminals.
package term

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

// Roughly test the terminal class and related methods.
func TestTerminal(t *testing.T) {
	t.Run("nil non-terminal", func(t *testing.T) {
		term := NewTerminal(nil)
		if term != nil {
			t.Error("NewTerminal(nil) didn't return nil")
		}

		gotFile := term.File()
		if gotFile != nil {
			t.Error("File() did not return nil")
		}

		gotTerminal := term.IsTerminal()
		if gotTerminal != false {
			t.Error("IsTerminal() is not false")
		}

		_, gotSizeErr := term.GetSize()
		if gotSizeErr != ErrNotATerminal {
			t.Error("GetSize() is not ErrNotATerminal")
		}

		gotResize := term.ResizeTo(WindowSize{})
		if gotResize != ErrNotATerminal {
			t.Error("ResizeTo() is not ErrNotATerminal")
		}

		gotClose := term.Close()
		if gotClose != nil {
			t.Error("Close() did not return nil")
		}
	})

	t.Run("file non-terminal", func(t *testing.T) {
		file, err := ioutil.TempFile("", "dummy")
		if err != nil {
			t.Fatal("TempFile failed to create")
		}
		defer os.Remove(file.Name())

		term := NewTerminal(file)
		if term == nil {
			t.Error("NewTerminal(file) returned nil")
		}

		gotFile := term.File()
		if gotFile == nil {
			t.Error("File() returned nil")
		}

		gotTerminal := term.IsTerminal()
		if gotTerminal != false {
			t.Error("IsTerminal() is not false")
		}

		_, gotSizeErr := term.GetSize()
		if gotSizeErr != ErrNotATerminal {
			t.Error("GetSize() is not ErrNotATerminal")
		}

		gotResize := term.ResizeTo(WindowSize{})
		if gotResize != ErrNotATerminal {
			t.Error("ResizeTo() is not ErrNotATerminal")
		}

		gotClose := term.Close()
		if gotClose != nil {
			t.Error("Close() did not return nil")
		}

	})

	t.Run("real terminal", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Real terminal not supported on windows")
		}
		file, err := ioutil.TempFile("", "dummy")
		if err != nil {
			t.Fatal("TempFile failed to create")
		}
		defer os.Remove(file.Name())

		term, _, err := OpenTerminal()
		if err != nil {
			t.Error("OpenTerminal() returned error")
		}
		defer term.Close()

		gotFile := term.File()
		if gotFile == nil {
			t.Error("File() did not return correct file")
		}

		gotTerminal := term.IsTerminal()
		if gotTerminal != true {
			t.Error("IsTerminal() is not true")
		}

		gotSizeActual, gotSizeErr := term.GetSize()
		if gotSizeErr != nil || gotSizeActual == nil {
			t.Error("GetSize() did not return non-nil size")
		}

		gotResize := term.ResizeTo(*gotSizeActual)
		if gotResize != nil {
			t.Error("ResizeTo() is not nil")
		}

		gotClose := term.Close()
		if gotClose != nil {
			t.Error("Close() did not return nil")
		}

	})
}
