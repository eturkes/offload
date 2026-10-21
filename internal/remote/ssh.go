// Package remote builds the ssh + rsync invocations offload uses to reach pool hosts.
package remote

import (
	"context"
	"os/exec"
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

// SSHArgs = ssh argv (without "ssh") running script on address.
func SSHArgs(address, script string) []string {
	args := append([]string{}, SSHOptions...)
	return append(args, "-T", "--", address, script)
}

// RsyncShell = rsync's -e value: ssh with SSHOptions.
func RsyncShell() string {
	return "ssh " + strings.Join(SSHOptions, " ")
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
