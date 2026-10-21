// Package remote builds the ssh + rsync invocations offload uses to reach pool hosts.
package remote

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// SSHOptions apply to every ssh connection, rsync's included. BatchMode: an agent cannot answer a
// prompt. ServerAlive*: a dead link ends ssh in ~1 min instead of hanging a run forever.
var SSHOptions = []string{
	"-o", "BatchMode=yes",
	"-o", "ConnectTimeout=5",
	"-o", "StrictHostKeyChecking=accept-new",
	"-o", "ServerAliveInterval=15",
	"-o", "ServerAliveCountMax=4",
}

// Target = one pool host as ssh + rsync reach it: its address + options beyond SSHOptions.
type Target struct {
	Address string
	Opts    []string
}

// SSHArgs = ssh argv (without "ssh") running script on the target.
func (t Target) SSHArgs(script string) []string {
	args := append(append([]string{}, SSHOptions...), t.Opts...)
	return append(args, "-T", "--", t.Address, script)
}

// RsyncShell = rsync's -e value: ssh with SSHOptions + the target's options. rsync splits it on
// whitespace, so option values never hold any (Multiplex checks its path).
func (t Target) RsyncShell() string {
	return "ssh " + strings.Join(append(append(append([]string{}, SSHOptions...), t.Opts...), "-T"), " ")
}

// controlDir = characters ssh's -o + rsync's -e parsers both take unquoted, `%` (a token) excluded.
var controlDir = regexp.MustCompile(`^[A-Za-z0-9/._-]+$`)

// maxControlDir: a socket path holds ≤ 107 bytes, and OpenSSH binds `<dir>/<%C: 40 hex>` through a
// temporary `.<16 chars>` sibling first ⇒ dir ≤ 107 - 1 - 40 - 17.
const maxControlDir = 49

// Multiplex = the ssh options sharing one master connection per host (ControlPersist 60 s idle),
// sockets in $XDG_RUNTIME_DIR/offload/ssh (created 0700). An error = why it stays off.
func Multiplex(getenv func(string) string) ([]string, error) {
	rt := getenv("XDG_RUNTIME_DIR")
	if rt == "" {
		return nil, fmt.Errorf("XDG_RUNTIME_DIR is unset")
	}
	dir := filepath.Join(rt, "offload", "ssh")
	if !controlDir.MatchString(dir) || len(dir) > maxControlDir {
		return nil, fmt.Errorf("socket directory %q is not a short plain path", dir)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return []string{"-o", "ControlMaster=auto", "-o", "ControlPath=" + dir + "/%C", "-o", "ControlPersist=60"}, nil
}

// Quote = s as one POSIX shell word.
func Quote(s string) string {
	if s != "" && strings.Trim(s, safeChars) == "" {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Join = words as one shell command line.
func Join(words []string) string {
	q := make([]string, len(words))
	for i, w := range words {
		q[i] = Quote(w)
	}
	return strings.Join(q, " ")
}

const safeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@%+=:,./_-"

// Command = exec.CommandContext with a graceful stop: cancel sends SIGTERM (rsync/ssh clean up),
// SIGKILL after 30 s. Own process group ⇒ a terminal's Ctrl-C reaches offload alone, which then
// stops its children in order.
func Command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: callers pass fixed tools (rsync, ssh, bash)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 30 * time.Second
	return cmd
}
