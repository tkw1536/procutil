package term

import (
	"os"
	"os/exec"
	"testing"
)

// Test that ExecTerminal indeed opens a terminal.
// We use the 'tty' command to be sure.
func TestExecTerminal(t *testing.T) {

	if !PTYSupport {
		t.Skip("OS not supported")
	}

	tty, err := exec.LookPath("tty")
	if err != nil {
		t.Skip("tty not found in path")
	}

	// make a command
	cmd := &exec.Cmd{
		Path: tty,
	}

	// start it on a pty
	term, err := ExecTerminal(cmd)
	if err != nil {
		t.Error(err)
		t.Fatal("ExecPty() returned err != nil")
	}
	defer term.Close()
	if term == nil || !term.IsTerminal() {
		t.Fatal("ExecPty(): did not return a terminal")
	}

	// make sure that the exit code == 0
	if err := cmd.Wait(); err != nil {
		t.Fatal("ExecPty(): not a tty")
	}
}

func CheckIfTerminal() {
	os.Exit(func() int {
		term, _, _, cleanup, err := GetStdTerminal()
		defer cleanup()
		if err != nil {
			return 1 // error
		}

		if term == nil {
			return 2 // it's not a terminal
		}

		return 0 // it's an actual terminal
	}())
}

// Test that GetStdTerminal() returns valid information about a terminal.
// We spawn a subprocess, and have that perform the check using its' exit code.
func TestGetStdTerminal(t *testing.T) {
	if !PTYSupport {
		t.Skip("OS not supported")
	}

	if os.Getenv("BE_STDTERM") == "1" {
		CheckIfTerminal()
		return
	}

	t.Run("GetStdTerminal on non-terminal", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestGetStdTerminal")
		cmd.Env = append(os.Environ(), "BE_STDTERM=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && e.ProcessState.ExitCode() == 2 { // 2 === it's not a terminal
			return
		}
		t.Fatalf("GetStdTerminal(): expected status 2 but got %v", err)
	})

	t.Run("do not run on stdin", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestGetStdTerminal")
		cmd.Env = append(os.Environ(), "BE_STDTERM=1")
		pty, err := ExecTerminal(cmd)
		if err != nil {
			t.Fatal("ExecTerminal(): returned error")
		}
		defer pty.Close()

		err = cmd.Wait()
		if err == nil { // 0 === it's a terminal
			return
		}
		t.Fatalf("GetStdTerminal(): expected status 0 but got %v", err)
	})
}
