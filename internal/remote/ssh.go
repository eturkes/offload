// Package remote builds the ssh + rsync invocations offload uses to reach pool hosts.
package remote

import "strings"

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
