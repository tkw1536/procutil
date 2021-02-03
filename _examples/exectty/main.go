// Command exectty is a dummy command that starts '/bin/bash' on a new terminal.
// If there is no terminal, it doesn't do anything.
package main

import (
	"context"
	"os"

	"github.com/tkw1536/procutil"
	"github.com/tkw1536/procutil/term"
)

func main() {
	os.Exit(run())
}

func run() int {
	fd, TERM, resize, cleanup, err := term.GetStdTerminal()
	if err != nil || fd == nil {
		panic("Std: Not a terminal")
	}
	defer cleanup()

	cmd := procutil.Command{
		Process: &procutil.ExecProcess{
			Command: "/bin/bash",
		},
	}

	if err := cmd.Init(context.Background(), true); err != nil {
		panic(err)
	}

	if err := cmd.StartPty(fd.File(), TERM, resize); err != nil {
		panic(err)
	}
	defer cmd.Cleanup()

	code, err := cmd.Wait()
	if err != nil {
		panic(err)
	}
	return code
}
