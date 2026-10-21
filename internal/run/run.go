// Package run executes a command under offload's watcher, locally or on a pool host, and returns
// the command's exit status.
package run

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/eturkes/offload/internal/remote"
)

var (
	//go:embed watch.bash
	watchScript string
	//go:embed local.bash
	localScript string
	//go:embed remote.bash
	remoteScript string
)

// ErrNotStarted = the remote side failed before the command ran (no status record).
var ErrNotStarted = errors.New("remote setup failed")

// Local runs argv in the current directory under the watcher. lock (fd 3 in the watcher) stays
// held until the watcher's teardown ends, even if offload dies first (S5).
func Local(ctx context.Context, argv []string, lock *os.File) (int, error) {
	cmd := remote.Command(context.Background(), "bash", append([]string{"-c", watchScript + localScript, "offload"}, argv...)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if lock != nil {
		cmd.ExtraFiles = []*os.File{lock}
	}
	return supervise(ctx, cmd, nil)
}

// Remote = one remote run (Y5, X1, X6).
type Remote struct {
	Address  string
	Workdir  string // relative to the remote $HOME
	Meta     string
	Sub      string // cwd relative to the workdir root
	Mode     string // git | plain | nosync
	Paths    []string
	Excludes string
	Argv     []string
}

// Run executes r.Argv remotely. A nil error = the command ran; its status is returned unchanged.
func (r Remote) Run(ctx context.Context) (int, error) {
	nonce := rand.Text()
	words := []string{
		"bash", "-c", watchScript + remoteScript, "offload", r.Workdir, r.Meta, r.Sub, r.Mode,
		nonce, strconv.Itoa(len(r.Paths)), r.Excludes, remote.Join(r.Argv),
	}
	var preamble []byte
	if r.Mode == "git" {
		var b bytes.Buffer
		b.WriteString(nonce)
		b.WriteByte(0)
		for _, p := range r.Paths {
			b.WriteString(p)
			b.WriteByte(0)
		}
		preamble = b.Bytes()
	}
	cmd := remote.Command(context.Background(), "ssh", remote.SSHArgs(r.Address, remote.Join(words))...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	rc, err := supervise(ctx, cmd, preamble)
	if err != nil || ctx.Err() != nil || (rc != 125 && rc != 255) {
		return rc, err
	}
	// 125/255 = the command's own status, or the remote script / ssh failing: the status record
	// written after the command ran tells them apart (X6).
	out, err := exec.Command("ssh", remote.SSHArgs(r.Address, "cat "+remote.Quote(r.Meta+"/status"))...).Output() //nolint:gosec // G204: trusted pool address + quoted path
	if err == nil && strings.TrimSpace(string(out)) == nonce+" "+strconv.Itoa(rc) {
		return rc, nil
	}
	return rc, ErrNotStarted
}

// supervise starts cmd with offload's pipe as stdin, writes preamble, and closes the pipe when ctx
// ends (signal) so the watcher tears the command down; it returns the shell-style exit status.
func supervise(ctx context.Context, cmd *exec.Cmd, preamble []byte) (int, error) {
	in, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	if err = cmd.Start(); err != nil {
		return 0, err
	}
	done := make(chan struct{})
	go func() {
		if len(preamble) > 0 {
			_, _ = in.Write(preamble) // a closed pipe (remote gone) shows up in the exit status
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
		case <-done:
		}
		_ = in.Close()
	}()
	err = cmd.Wait()
	close(done)
	ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if !ok {
		return 0, fmt.Errorf("%s: %w", cmd.Path, err)
	}
	if ws.Signaled() {
		return 128 + int(ws.Signal()), nil
	}
	return ws.ExitStatus(), nil
}
