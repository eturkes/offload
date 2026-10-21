// Package cli parses offload's command line and dispatches its subcommands.
package cli

import (
	"fmt"
	"io"
)

// Exit statuses offload owns; every other status is the command's.
const (
	ExitUsage = 2
)

const usage = `Usage:
  offload run [--local | --pool] [--host NAME] [--gpu] [--no-sync] [--get PATH]... -- COMMAND [ARG...]
  offload status
`

// Main runs offload with args (without the program name) and returns the exit status.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(stdout, usage)
		return 0
	}
	fmt.Fprint(stderr, usage)
	return ExitUsage
}
