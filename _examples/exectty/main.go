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
	fd, TERM, resize, cleanup := term.GetStdTerminal()
	if fd == nil {
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

	code, err := cmd.Wait()
	if err != nil {
		panic(err)
	}
	return code
}
