package procutil

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/tkw1536/procutil/term"
)

// Parts of the code in this file is roughly adapted from https://github.com/docker/cli/blob/master/cli/command/container/exec.go
// and also https://github.com/docker/cli/blob/master/cli/command/container/hijack.go.
//
// These are licensed under the Apache 2.0 License.
// This license requires to state changes made to the code and inclusion of the original NOTICE file.
//
// The code was modified to be independent of the docker cli utility classes where applicable.
//
// The original license and NOTICE can be found below:
//
// Docker
// Copyright 2012-2017 Docker, Inc.
//
// This product includes software developed at Docker, Inc. (https://www.docker.com).
//
// This product contains software (https://github.com/creack/pty) developed
// by Keith Rarick, licensed under the MIT License.
//
// The following is courtesy of our legal counsel:
//
// Use and transfer of Docker may be subject to certain restrictions by the
// United States and other governments.
// It is your responsibility to ensure that your use and/or transfer does not
// violate applicable laws.
//
// For more information, please see https://www.bis.doc.gov
//
// See also https://www.apache.org/dev/crypto.html and/or seek legal counsel.

// NewDockerExecProcess creates a process that executes within a docker container.
func NewDockerExecProcess(client client.APIClient, containerID string, command []string) *StreamingProcess {
	return &StreamingProcess{
		Streamer: &DockerExecStreamer{
			client:      client,
			containerID: containerID,
			config: types.ExecConfig{
				AttachStdin:  true,
				AttachStderr: true,
				AttachStdout: true,
				Cmd:          command,
			},
		},
	}
}

// DockerExecStreamer is a streamer that streams data to and from a remote docker exec process
type DockerExecStreamer struct {
	// paramters
	client      client.APIClient
	containerID string
	config      types.ExecConfig

	// state
	execID string
	conn   *types.HijackedResponse
}

func (des *DockerExecStreamer) String() string {
	return strings.Join(append([]string{des.containerID}, des.config.Cmd...), " ")
}

// Init initializes this docker exec streamer
func (des *DockerExecStreamer) Init(ctx context.Context, Term string, isPty bool) error {
	if isPty {
		des.config.Tty = true
		des.config.Env = append(des.config.Env, "TERM="+Term)
	}
	return nil
}

// Attach attaches to this DockerExecStreamer
func (des *DockerExecStreamer) Attach(ctx context.Context, isPty bool) error {
	// create the exec
	res, err := des.client.ContainerExecCreate(ctx, des.containerID, des.config)
	if err != nil {
		return err
	}
	des.execID = res.ID

	// attach to it
	conn, err := des.client.ContainerExecAttach(ctx, des.execID, types.ExecStartCheck{
		Detach: false,
		Tty:    isPty,
	})
	if err != nil {
		return err
	}
	des.conn = &conn

	return nil
}

// ResizeTo resizes the remote stream
func (des *DockerExecStreamer) ResizeTo(ctx context.Context, size term.WindowSize) error {
	return des.client.ContainerExecResize(ctx, des.execID, types.ResizeOptions{
		Height: uint(size.Height),
		Width:  uint(size.Width),
	})
}

// Result returns the result of the stream
func (des *DockerExecStreamer) Result(ctx context.Context) (int, error) {
	res, err := des.client.ContainerExecInspect(ctx, des.execID)
	return res.ExitCode, err
}

// Detach detaches from the stream
func (des *DockerExecStreamer) Detach(ctx context.Context) error {
	des.conn.Close()
	return nil
}

// StreamOutput streams output from the remote stream
func (des *DockerExecStreamer) StreamOutput(ctx context.Context, stdout, stderr io.Writer, restoreTerms func(), errChan chan error) {
	var err error
	if stderr == nil {
		_, err = io.Copy(stdout, des.conn.Reader)
		restoreTerms()
	} else {
		_, err = stdcopy.StdCopy(stdout, stderr, des.conn.Reader)
	}
	errChan <- err
}

// StreamInput streams input to the remote stream
func (des *DockerExecStreamer) StreamInput(ctx context.Context, stdin io.Reader, restoreTerms func(), doneChan chan struct{}) {
	io.Copy(des.conn.Conn, stdin)
	des.conn.CloseWrite()
	close(doneChan)
}
