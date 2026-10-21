// Package run executes a command under offload's watcher, locally or on a pool host, and returns
// the command's exit status.
package run

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
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

	runScript = watchScript + remoteScript
	runSum    = func() string { h := sha256.Sum256([]byte(runScript)); return hex.EncodeToString(h[:]) }()
)

// boot = the ssh command's script (X8): $1 bytes of stdin (head -c never reads past them), sha256
// $2 (a short or corrupt copy fails it), become
// <meta>/run-<16 hex>.bash (temp file + rename; content-addressed ⇒ concurrent runs of other
// offload versions never swap it under each other), then that file runs — a remote ps shows its path,
// never the script. A short or corrupt copy never runs: 125. Its own failure path: the script's
// `fail` does not exist yet.
const boot = `__f=$4/run-${2:0:16}.bash
cd && mkdir -p "$4" && head -c "$1" >"$__f.$$" && [ "$(sha256sum <"$__f.$$")" = "$2  -" ] &&
  mv -f "$__f.$$" "$__f" || {
  rm -f "$__f.$$"
  printf 'offload: %s: cannot install the run script in %s\n' "$(uname -n)" "$4" >&2
  exit 125
}
shift 2
exec bash "$__f" "$@"`

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
	Target   remote.Target
	Workdir  string // relative to the remote $HOME
	Meta     string
	Sub      string // cwd relative to the workdir root
	Mode     string // git | worktree:<id> | plain | nosync
	Paths    []string
	Excludes string
	Argv     []string
}

// command = the ssh command line + the stdin head for a run with this nonce: the run script, then
// (git mode) NONCE, EXCLUDES and every upload path, NUL-terminated (X8, Y5).
func (r Remote) command(nonce string) (string, []byte) {
	words := []string{
		"bash", "-c", boot, "offload", strconv.Itoa(len(runScript)), runSum,
		r.Workdir, r.Meta, r.Sub, r.Mode, nonce, strconv.Itoa(len(r.Paths)), remote.Join(r.Argv),
	}
	var b bytes.Buffer
	b.WriteString(runScript)
	if r.Mode == "git" || strings.HasPrefix(r.Mode, "worktree:") {
		for _, rec := range append([]string{nonce, r.Excludes}, r.Paths...) {
			b.WriteString(rec)
			b.WriteByte(0)
		}
	}
	return remote.Join(words), b.Bytes()
}

// Run executes r.Argv remotely. A nil error = the command ran; its status is returned unchanged.
func (r Remote) Run(ctx context.Context) (int, error) {
	nonce := rand.Text()
	script, preamble := r.command(nonce)
	cmd := remote.Command(context.Background(), "ssh", r.Target.SSHArgs(script)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	rc, err := supervise(ctx, cmd, preamble)
	if err != nil || ctx.Err() != nil || (rc != 125 && rc != 255) {
		return rc, err
	}
	// 125/255 = the command's own status, or the remote script / ssh failing: the status record
	// written after the command ran tells them apart (X6).
	out, err := remote.Command(ctx, "ssh", r.Target.SSHArgs("cat "+remote.Quote(r.Meta+"/status"))...).Output()
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
